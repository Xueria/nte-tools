package dev.xueria.nte.tools.music

data class MusicTimeSignature(
    val top: Int,
    val bottom: Int
)

data class MusicNote(
    val octave: Int,
    val degree: Int,
    val accidental: Int,
    val beats: Float
)

data class MusicTrack(
    val id: Int,
    val notes: List<MusicNote>
)

data class MusicSheet(
    val name: String,
    val timeSignature: MusicTimeSignature,
    val keySignature: KeySignature,
    val bpm: Float,
    val tracks: List<MusicTrack>
)