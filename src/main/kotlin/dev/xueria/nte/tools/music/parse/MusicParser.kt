package dev.xueria.nte.tools.music.parse

import dev.xueria.nte.tools.music.MusicSheet

interface MusicParser {

    fun supportType(): String

    fun parse(input: ParseInput): Result<MusicSheet>

}