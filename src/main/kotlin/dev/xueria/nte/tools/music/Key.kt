package dev.xueria.nte.tools.music

enum class Tone(val symbol: String, val semitone: Int) {
    C(symbol = "C", semitone = 0),
    C_SHARP(symbol = "C#", semitone = 1),
    D(symbol = "D", semitone = 2),
    E_FLAT(symbol = "Eb", semitone = 3),
    E(symbol = "E", semitone = 4),
    F(symbol = "F", semitone = 5),
    F_SHARP(symbol = "F#", semitone = 6),
    G(symbol = "G", semitone = 7),
    G_SHARP(symbol = "G#", semitone = 8),
    A(symbol = "A", semitone = 9),
    B_FLAT(symbol = "Bb", semitone = 10),
    B(symbol = "B", semitone = 11)
}

@JvmInline
value class Pitch(val value: Int) : Comparable<Pitch> {

    companion object {

        val C4: Pitch = Pitch(value = 60)

    }

    override fun compareTo(other: Pitch): Int = value.compareTo(other.value)
}

data class KeySignature(val tonic: Tone) {

    companion object {

        val C_MAJOR: KeySignature = KeySignature(Tone.C)

        private val WHITE_STEPS: IntArray = intArrayOf(0, 2, 4, 5, 7, 9, 11)

        private const val OCTAVE_STEPS: Int = 12

    }

    fun pitchOf(key: MusicNote): Pitch {
        val pitch = Pitch.C4.value + tonic.semitone +
                OCTAVE_STEPS * key.octave +
                WHITE_STEPS[key.degree - 1] + key.accidental

        require(pitch in 1..127) {
            "invalid pitch"
        }

        return Pitch(pitch)
    }

    fun pitchOf(octave: Int, degree: Int, accidental: Int = 0): Pitch =
        Pitch(
            value = Pitch.C4.value +
                    tonic.semitone +
                    OCTAVE_STEPS * octave +
                    WHITE_STEPS[degree - 1] +
                    accidental
        )


}