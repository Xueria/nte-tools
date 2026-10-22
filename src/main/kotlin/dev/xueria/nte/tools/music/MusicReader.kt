package dev.xueria.nte.tools.music

interface MusicReader {

    fun read(name: String, source: String): MusicSheet

}