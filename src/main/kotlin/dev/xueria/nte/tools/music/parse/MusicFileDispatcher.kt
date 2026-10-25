package dev.xueria.nte.tools.music.parse

import dev.xueria.nte.tools.music.MusicSheet

object MusicFileDispatcher {

    private val parsers: MutableMap<String, MusicParser> = mutableMapOf()

    fun buildIn() {
        register(MidiMusicParser)
    }

    fun register(parser: MusicParser) {
        parsers[parser.supportType()] = parser
    }

    fun dispatch(input: ParseInput): Result<MusicSheet> {
        val parser = parsers[input.type]
            ?: throw Throwable("Not Support type: ${input.type}")

        return parser.parse(input)
    }


}