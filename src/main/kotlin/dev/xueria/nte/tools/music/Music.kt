package dev.xueria.nte.tools.music

data class TimeSignature(
    val beatsPerBar: Int,
    val beatUnit: Int
)

data class MusicNote(
    val tone: String,
    val durationBeats: Int
)

data class MusicSheet(
    val name: String,
    val key: String,
)