package dev.xueria.nte.tools.music.parse

import dev.xueria.nte.tools.music.MusicSheet
import dev.xueria.nte.tools.music.parse.midi.MidiFileReader
import dev.xueria.nte.tools.music.parse.midi.MidiSheetMapper

/**
 * Standard MIDI File parser backed by [javax.sound.midi.MidiSystem].
 *
 * Supported: formats 0, 1 and 2, running status, meta events (track name,
 * tempo, time signature, key signature), both PPQ and SMPTE time divisions.
 * The first tempo, time signature and key signature of the file win; later
 * changes are ignored because the sheet model cannot express them.
 */
object MidiMusicParser : MusicParser {

    private const val TYPE: String = "midi"

    override fun supportType(): String = TYPE

    override fun parse(input: ParseInput): Result<MusicSheet> = runCatching {
        MidiSheetMapper.toSheet(MidiFileReader.read(input.bytes))
    }

}