package dev.xueria.nte.tools.music.play

import dev.xueria.nte.tools.music.KeyMapper
import dev.xueria.nte.tools.music.KeySignature
import dev.xueria.nte.tools.music.MusicNote
import dev.xueria.nte.tools.music.MusicSheet
import dev.xueria.nte.tools.music.MusicTrack
import dev.xueria.nte.tools.music.Pitch
import kotlin.math.abs

/**
 * Folds the notes of a [dev.xueria.nte.tools.music.MusicSheet] into the pitch range covered by a [dev.xueria.nte.tools.music.KeyMapper].
 *
 * A note keeps its scale degree and accidental; only its octave is changed, so
 * the melody stays in key and every produced note is mapped by the keymap.
 *
 * A pitch class that the keymap does not contain at any octave cannot be played
 * at all (for example a sharp on a diatonic-only instrument). Such a note is
 * dropped and its length is added to the previous note, so the remaining notes
 * keep their original timing.
 */
class MusicSheetCompressor(keyMapper: KeyMapper) {

    private val playablePitches: List<Pitch> = (MIN_PITCH..MAX_PITCH)
        .map { value -> Pitch(value) }
        .filter { pitch -> keyMapper.keyof(pitch) != null }

    private val playableByTone: Map<Int, List<Pitch>> =
        playablePitches.groupBy { pitch -> Math.floorMod(pitch.value, SEMITONES_PER_OCTAVE) }

    init {
        require(playablePitches.isNotEmpty()) { "Keymap contains no playable pitch" }
    }

    /** Returns a copy of [sheet] that only uses pitches present in the keymap. */
    fun compress(sheet: MusicSheet): MusicSheet = sheet.copy(
        tracks = sheet.tracks.map { track -> fit(track, sheet.keySignature) }
    )

    private fun fit(track: MusicTrack, keySignature: KeySignature): MusicTrack {
        val fitted: MutableList<MusicNote> = mutableListOf()

        track.notes.forEach { note -> add(note, fitted, keySignature) }

        return track.copy(notes = fitted)
    }

    private fun add(note: MusicNote, fitted: MutableList<MusicNote>, keySignature: KeySignature) {
        val folded: MusicNote? = fold(note, keySignature)

        if (folded != null) {
            fitted += folded
            return
        }

        extend(fitted, note.beats)
    }

    /** Returns [note] moved by whole octaves onto the nearest playable pitch. */
    private fun fold(note: MusicNote, keySignature: KeySignature): MusicNote? {
        val pitch: Pitch = keySignature.pitchOf(note.octave, note.degree, note.accidental)
        val target: Pitch = nearest(pitch) ?: return null
        val octaves: Int = (target.value - pitch.value) / SEMITONES_PER_OCTAVE

        return note.copy(octave = note.octave + octaves)
    }

    private fun nearest(pitch: Pitch): Pitch? {
        val candidates: List<Pitch> =
            playableByTone[Math.floorMod(pitch.value, SEMITONES_PER_OCTAVE)] ?: return null
        return candidates.minByOrNull { candidate -> abs(candidate.value - pitch.value) }
    }

    private fun extend(notes: MutableList<MusicNote>, beats: Float) {
        val last: MusicNote = notes.lastOrNull() ?: return
        notes[notes.lastIndex] = last.copy(beats = last.beats + beats)
    }

    private companion object {
        const val MIN_PITCH: Int = 0
        const val MAX_PITCH: Int = 127
        const val SEMITONES_PER_OCTAVE: Int = 12
    }

}