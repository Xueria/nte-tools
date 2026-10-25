package dev.xueria.nte.tools.music.parse.midi

import java.io.ByteArrayInputStream
import java.io.IOException
import javax.sound.midi.InvalidMidiDataException
import javax.sound.midi.MetaMessage
import javax.sound.midi.MidiEvent
import javax.sound.midi.MidiMessage
import javax.sound.midi.MidiSystem
import javax.sound.midi.Sequence
import javax.sound.midi.ShortMessage
import javax.sound.midi.Track

/**
 * Reads a Standard MIDI File with [MidiSystem] and extracts the information the
 * sheet model needs.
 */
internal object MidiFileReader {

    private const val DEFAULT_MICROSECONDS_PER_QUARTER: Int = 500_000

    fun read(bytes: ByteArray): MidiSong {
        val sequence: Sequence = load(bytes)
        val tracks: List<MidiTrack> = sequence.tracks.mapIndexed { index, track -> readTrack(index, track) }

        return MidiSong(
            name = tracks.firstNotNullOfOrNull { track -> track.name?.takeIf { name -> name.isNotBlank() } },
            divisionType = sequence.divisionType,
            resolution = sequence.resolution,
            microsecondsPerQuarter = firstTempo(tracks)?.microsecondsPerQuarter
                ?: DEFAULT_MICROSECONDS_PER_QUARTER,
            timeSignature = firstTimeSignature(tracks),
            keySignature = firstKeySignature(tracks),
            notes = tracks.flatMap { track -> track.notes }
        )
    }

    private fun load(bytes: ByteArray): Sequence = try {
        MidiSystem.getSequence(ByteArrayInputStream(bytes))
    } catch (exception: InvalidMidiDataException) {
        throw MidiFormatException("Invalid MIDI data: ${exception.message ?: "unknown format"}", exception)
    } catch (exception: IOException) {
        throw MidiFormatException("Cannot read MIDI data: ${exception.message ?: "I/O error"}", exception)
    }

    private fun readTrack(index: Int, track: Track): MidiTrack {
        val collector: TrackCollector = TrackCollector(index)

        for (eventIndex in 0 until track.size()) {
            collector.accept(track.get(eventIndex))
        }

        return collector.build()
    }

    private fun firstTempo(tracks: List<MidiTrack>): MidiTempo? =
        tracks.flatMap { track -> track.tempos }.minByOrNull { tempo -> tempo.tick }

    private fun firstTimeSignature(tracks: List<MidiTrack>): MidiTimeSignature? =
        tracks.flatMap { track -> track.timeSignatures }.minByOrNull { signature -> signature.tick }

    private fun firstKeySignature(tracks: List<MidiTrack>): MidiKeySignature? =
        tracks.flatMap { track -> track.keySignatures }.minByOrNull { signature -> signature.tick }

}

/**
 * Collects the notes and meta events of one track.
 *
 * Note-on and note-off events are paired per channel and pitch, so repeated or
 * nested notes of the same pitch stay correctly separated.
 */
private class TrackCollector(private val track: Int) {

    private val notes: MutableList<MidiNote> = mutableListOf()
    private val tempos: MutableList<MidiTempo> = mutableListOf()
    private val timeSignatures: MutableList<MidiTimeSignature> = mutableListOf()
    private val keySignatures: MutableList<MidiKeySignature> = mutableListOf()
    private val openNotes: MutableMap<NoteKey, ArrayDeque<Long>> = mutableMapOf()

    private var name: String? = null

    fun accept(event: MidiEvent) {
        val message: MidiMessage = event.message

        when (message) {
            is ShortMessage -> acceptShort(event.tick, message)
            is MetaMessage -> acceptMeta(event.tick, message)
            else -> Unit
        }
    }

    fun build(): MidiTrack = MidiTrack(track, name, tempos, timeSignatures, keySignatures, notes)

    private fun acceptShort(tick: Long, message: ShortMessage) {
        val pitch: Int = message.data1

        if (message.command == ShortMessage.NOTE_OFF) {
            close(tick, message.channel, pitch)
            return
        }

        if (message.command != ShortMessage.NOTE_ON) return

        if (message.data2 == SILENT_VELOCITY) {
            close(tick, message.channel, pitch)
            return
        }

        open(tick, message.channel, pitch)
    }

    private fun acceptMeta(tick: Long, message: MetaMessage) {
        val data: ByteArray = message.data

        when (message.type) {
            META_TRACK_NAME -> name = decodeText(data)
            META_TEMPO -> recordTempo(tick, data)
            META_TIME_SIGNATURE -> recordTimeSignature(tick, data)
            META_KEY_SIGNATURE -> recordKeySignature(tick, data)
            else -> Unit
        }
    }

    private fun open(tick: Long, channel: Int, pitch: Int) {
        openNotes.getOrPut(NoteKey(channel, pitch)) { ArrayDeque() }.addLast(tick)
    }

    private fun close(tick: Long, channel: Int, pitch: Int) {
        val starts: ArrayDeque<Long> = openNotes[NoteKey(channel, pitch)] ?: return
        val start: Long = starts.removeFirstOrNull() ?: return

        if (tick > start) notes += MidiNote(track, channel, pitch, start, tick - start)
    }

    private fun recordTempo(tick: Long, data: ByteArray) {
        if (data.size < TEMPO_BYTES) return

        val microseconds: Int = (unsigned(data, 0) shl 16) or (unsigned(data, 1) shl 8) or unsigned(data, 2)

        if (microseconds > 0) tempos += MidiTempo(tick, microseconds)
    }

    private fun recordTimeSignature(tick: Long, data: ByteArray) {
        if (data.size < TIME_SIGNATURE_BYTES) return

        val top: Int = unsigned(data, 0)
        val power: Int = unsigned(data, 1).coerceIn(0, MAX_DENOMINATOR_POWER)

        if (top > 0) timeSignatures += MidiTimeSignature(tick, top, 1 shl power)
    }

    private fun recordKeySignature(tick: Long, data: ByteArray) {
        if (data.size < KEY_SIGNATURE_BYTES) return

        val sharps: Int = data[0].toInt()

        if (sharps in MIN_SHARPS..MAX_SHARPS) {
            keySignatures += MidiKeySignature(tick, sharps, data[1].toInt() != 0)
        }
    }

    private fun unsigned(data: ByteArray, index: Int): Int = data[index].toInt() and BYTE_MASK

    private companion object {
        const val SILENT_VELOCITY: Int = 0
        const val META_TRACK_NAME: Int = 0x03
        const val META_TEMPO: Int = 0x51
        const val META_TIME_SIGNATURE: Int = 0x58
        const val META_KEY_SIGNATURE: Int = 0x59
        const val TEMPO_BYTES: Int = 3
        const val TIME_SIGNATURE_BYTES: Int = 4
        const val KEY_SIGNATURE_BYTES: Int = 2
        const val MAX_DENOMINATOR_POWER: Int = 7
        const val MIN_SHARPS: Int = -7
        const val MAX_SHARPS: Int = 7
        const val BYTE_MASK: Int = 0xFF
    }

}

private data class NoteKey(val channel: Int, val pitch: Int)

/** Track names are usually ASCII, but modern files may store UTF-8. */
private fun decodeText(data: ByteArray): String {
    val utf8: String = String(data, Charsets.UTF_8)
    val text: String = if ('\uFFFD' in utf8) String(data, Charsets.ISO_8859_1) else utf8
    return text.trim()
}
