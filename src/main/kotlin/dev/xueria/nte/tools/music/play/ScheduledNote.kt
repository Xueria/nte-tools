package dev.xueria.nte.tools.music.play

import dev.xueria.nte.tools.music.KeyMapper
import dev.xueria.nte.tools.music.KeySignature
import dev.xueria.nte.tools.music.KeyStroke
import dev.xueria.nte.tools.music.MusicSheet
import dev.xueria.nte.tools.music.MusicTrack
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

/** One playable note of a [MusicSchedule]. */
data class ScheduledNote(
    val begin: Duration,
    val stroke: KeyStroke
)

/**
 * A [MusicSheet] flattened into absolute press times.
 *
 * Tracks play in parallel, while the notes of one track are sequential: a note
 * starts after the previous note of the same track has used up its beats.
 */
data class MusicSchedule(
    val notes: List<ScheduledNote>,
    val duration: Duration,
    val unplayable: Int
) {

    companion object {

        private const val SECONDS_PER_MINUTE: Double = 60.0

        /** Builds a schedule; notes without a key binding are only counted in [unplayable]. */
        fun of(sheet: MusicSheet, keyMapper: KeyMapper): MusicSchedule {
            val secondsPerBeat: Double = SECONDS_PER_MINUTE / sheet.bpm
            val builder: ScheduleBuilder = ScheduleBuilder(secondsPerBeat)

            sheet.tracks.forEach { track -> builder.add(track, sheet.keySignature, keyMapper) }

            return builder.build()
        }

    }

}

private class ScheduleBuilder(private val secondsPerBeat: Double) {

    private val notes: MutableList<ScheduledNote> = mutableListOf()

    private var unplayable: Int = 0
    private var duration: Duration = Duration.ZERO

    fun add(track: MusicTrack, keySignature: KeySignature, keyMapper: KeyMapper) {
        var beats: Double = 0.0

        track.notes.forEach { note ->
            val pitch = keySignature.pitchOf(note.octave, note.degree, note.accidental)
            val stroke: KeyStroke? = keyMapper.keyof(pitch)

            if (stroke == null) unplayable++ else notes += ScheduledNote(secondsOf(beats), stroke)

            beats += note.beats
        }

        duration = maxOf(duration, secondsOf(beats))
    }

    fun build(): MusicSchedule =
        MusicSchedule(notes.sortedBy { note -> note.begin }, duration, unplayable)

    private fun secondsOf(beats: Double): Duration = (beats * secondsPerBeat).seconds

}