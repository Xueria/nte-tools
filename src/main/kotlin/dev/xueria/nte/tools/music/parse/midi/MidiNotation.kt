package dev.xueria.nte.tools.music.parse.midi

import dev.xueria.nte.tools.music.KeySignature
import dev.xueria.nte.tools.music.MusicNote
import dev.xueria.nte.tools.music.Pitch
import dev.xueria.nte.tools.music.Tone

/**
 * Maps a MIDI key signature to the major tonic used by [KeySignature].
 *
 * MIDI stores the number of sharps (positive) or flats (negative). Minor keys
 * share that signature with their relative major, so the relative major tonic
 * is used for both modes. [Tone] only holds twelve spellings, therefore
 * enharmonic keys fall back to the closest available spelling (Cb to B, Gb to
 * F#, Db to C#, Ab to G#).
 */
internal object MidiKeySignatures {

    private val MAJOR_TONICS: Map<Int, Tone> = mapOf(
        -7 to Tone.B,
        -6 to Tone.F_SHARP,
        -5 to Tone.C_SHARP,
        -4 to Tone.G_SHARP,
        -3 to Tone.E_FLAT,
        -2 to Tone.B_FLAT,
        -1 to Tone.F,
        0 to Tone.C,
        1 to Tone.G,
        2 to Tone.D,
        3 to Tone.A,
        4 to Tone.E,
        5 to Tone.B,
        6 to Tone.F_SHARP,
        7 to Tone.C_SHARP
    )

    fun of(signature: MidiKeySignature?): KeySignature {
        val sharps: Int = signature?.sharps ?: 0
        return KeySignature(MAJOR_TONICS[sharps] ?: Tone.C)
    }

}

/** A scale degree plus its chromatic alteration. */
private data class DegreeNotation(val degree: Int, val accidental: Int)

/**
 * Converts absolute MIDI pitches into the degree notation used by [KeySignature].
 *
 * [KeySignature.pitchOf] anchors octave zero at the tonic inside the octave of
 * middle C, so this class mirrors that layout exactly and every produced note
 * round-trips back to the original pitch.
 */
internal class MidiPitchNotation private constructor(
    private val tonicSemitone: Int,
    private val degrees: List<DegreeNotation>
) {

    fun noteOf(pitch: Int, beats: Float): MusicNote {
        val relative: Int = pitch - Pitch.C4.value - tonicSemitone
        val octave: Int = Math.floorDiv(relative, SEMITONES_PER_OCTAVE)
        val degree: DegreeNotation = degrees[Math.floorMod(relative, SEMITONES_PER_OCTAVE)]

        return MusicNote(
            octave = octave,
            degree = degree.degree,
            accidental = degree.accidental,
            beats = beats
        )
    }

    companion object {

        private const val SEMITONES_PER_OCTAVE: Int = 12
        private val MAJOR_SCALE_SEMITONES: List<Int> = listOf(0, 2, 4, 5, 7, 9, 11)

        fun of(keySignature: KeySignature, signature: MidiKeySignature?): MidiPitchNotation {
            val prefersSharps: Boolean = (signature?.sharps ?: 0) >= 0

            return MidiPitchNotation(
                tonicSemitone = keySignature.tonic.semitone,
                degrees = (0 until SEMITONES_PER_OCTAVE).map { semitone -> degreeOf(semitone, prefersSharps) }
            )
        }

        private fun degreeOf(semitone: Int, prefersSharps: Boolean): DegreeNotation {
            val naturalDegree: Int = MAJOR_SCALE_SEMITONES.indexOf(semitone)
            if (naturalDegree >= 0) return DegreeNotation(naturalDegree + 1, 0)

            val anchor: Int = if (prefersSharps) semitone - 1 else semitone + 1
            val accidental: Int = if (prefersSharps) 1 else -1

            return DegreeNotation(MAJOR_SCALE_SEMITONES.indexOf(anchor) + 1, accidental)
        }

    }

}
