package dev.xueria.nte.tools.music.play

import dev.xueria.nte.tools.music.KeyMapper
import dev.xueria.nte.tools.music.MusicSheet
import kotlin.time.Duration

/**
 * A [MusicSheet] prepared for one [KeyMapper].
 *
 * Compression and scheduling happen once, when the plan is built, so the UI can
 * show the result before playing and the [AutoPlayer] only has to follow the
 * schedule.
 */
class PlaybackPlan(
    val sheet: MusicSheet,
    val keyMapper: KeyMapper
) {

    /** The sheet folded into the pitch range of [keyMapper]. */
    val playable: MusicSheet = MusicSheetCompressor(keyMapper).compress(sheet)

    val schedule: MusicSchedule = MusicSchedule.of(playable, keyMapper)

    val duration: Duration = schedule.duration

    val totalNotes: Int = sheet.tracks.sumOf { track -> track.notes.size }

    val playableNotes: Int = playable.tracks.sumOf { track -> track.notes.size }

    /** Notes that no octave of the keymap can play. */
    val removedNotes: Int
        get() = totalNotes - playableNotes

}