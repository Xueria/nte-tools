package dev.xueria.nte.tools.music.parse.midi

import dev.xueria.nte.tools.music.KeySignature
import dev.xueria.nte.tools.music.MusicNote
import dev.xueria.nte.tools.music.MusicSheet
import dev.xueria.nte.tools.music.MusicTimeSignature
import dev.xueria.nte.tools.music.MusicTrack

/** Identifies one MIDI part: a track/channel pair. */
private data class MidiPart(val track: Int, val channel: Int)

private val PART_ORDER: Comparator<MidiPart> = compareBy({ part -> part.track }, { part -> part.channel })

/**
 * Converts a parsed [MidiSong] into the playable [MusicSheet].
 *
 * [MusicNote.beats] holds the interval to the next onset of the same voice, not
 * the sounding length: the sheet has no rests and no absolute time, so a player
 * that accumulates beats per track replays the original timeline exactly. The
 * last note of a voice falls back to its own length.
 */
internal object MidiSheetMapper {

    fun toSheet(song: MidiSong): MusicSheet {
        val playable: List<MidiNote> = song.notes.filter(::isPlayable)

        if (playable.isEmpty()) throw MidiFormatException("MIDI file contains no playable notes")

        val keySignature: KeySignature = MidiKeySignatures.of(song.keySignature)
        val notation: MidiPitchNotation = MidiPitchNotation.of(keySignature, song.keySignature)
        val converter: MidiTickConverter =
            MidiTickConverter.of(song.divisionType, song.resolution, song.microsecondsPerQuarter)
        val voices: List<List<MidiNote>> = voicesOf(playable)

        return MusicSheet(
            name = song.name ?: DEFAULT_NAME,
            timeSignature = song.timeSignature?.asModel() ?: DEFAULT_TIME_SIGNATURE,
            keySignature = keySignature,
            bpm = bpmOf(song.microsecondsPerQuarter),
            tracks = voices.mapIndexed { index, voice -> trackOf(index + 1, voice, notation, converter) }
        )
    }

    private fun voicesOf(notes: List<MidiNote>): List<List<MidiNote>> =
        notes.groupBy { note -> MidiPart(note.track, note.channel) }
            .toSortedMap(PART_ORDER)
            .values
            .flatMap(MidiVoiceSplitter::split)

    private fun trackOf(
        id: Int,
        voice: List<MidiNote>,
        notation: MidiPitchNotation,
        converter: MidiTickConverter
    ): MusicTrack {
        val notes: List<MusicNote> = voice.mapIndexed { index, note ->
            notation.noteOf(note.pitch, beatsOf(voice, index, converter))
        }

        return MusicTrack(id = id, notes = notes)
    }

    private fun beatsOf(voice: List<MidiNote>, index: Int, converter: MidiTickConverter): Float {
        val note: MidiNote = voice[index]
        val next: MidiNote? = voice.getOrNull(index + 1)
        val ticks: Long = if (next == null) note.durationTicks else next.startTick - note.startTick

        return converter.beatsOf(ticks)
    }

    private fun isPlayable(note: MidiNote): Boolean =
        note.channel != PERCUSSION_CHANNEL && note.pitch in MIN_PITCH..MAX_PITCH

    private fun bpmOf(microsecondsPerQuarter: Int): Float =
        MICROSECONDS_PER_MINUTE / microsecondsPerQuarter

    private fun MidiTimeSignature.asModel(): MusicTimeSignature = MusicTimeSignature(top = top, bottom = bottom)

    private const val DEFAULT_NAME: String = "Untitled"
    private const val PERCUSSION_CHANNEL: Int = 9
    private const val MIN_PITCH: Int = 1
    private const val MAX_PITCH: Int = 127
    private const val MICROSECONDS_PER_MINUTE: Float = 60_000_000f

    private val DEFAULT_TIME_SIGNATURE: MusicTimeSignature = MusicTimeSignature(top = 4, bottom = 4)

}
