package dev.xueria.nte.tools.music.play

import dev.xueria.nte.tools.music.KeyStroke
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import java.util.concurrent.ConcurrentHashMap
import kotlin.time.Duration
import kotlin.time.Duration.Companion.milliseconds
import kotlin.time.Duration.Companion.nanoseconds
import kotlin.time.Duration.Companion.seconds

/**
 * What the [AutoPlayer] is doing right now.
 *
 * [playing] stays `true` while the playback is paused, so the UI can tell a
 * paused session from a stopped one.
 */
data class PlaybackState(
    val playing: Boolean = false,
    val paused: Boolean = false,
    val index: Int = -1,
    val stroke: KeyStroke? = null
)

/**
 * Plays a [MusicSchedule] by sending real key presses.
 *
 * [hold] is how long a key stays down; it is shortened when the same key has to
 * be pressed again sooner. [leadIn] gives the user time to focus the game window.
 */
class AutoPlayer(
    private val sender: KeySender,
    private val hold: Duration = DEFAULT_HOLD,
    private val leadIn: Duration = DEFAULT_LEAD_IN,
    private val scope: CoroutineScope = CoroutineScope(SupervisorJob() + Dispatchers.Default)
) {

    private val mutableState: MutableStateFlow<PlaybackState> = MutableStateFlow(PlaybackState())
    private val paused: MutableStateFlow<Boolean> = MutableStateFlow(false)
    private val pressedKeys: MutableSet<KeyStroke> = ConcurrentHashMap.newKeySet()

    private var job: Job? = null

    @Volatile
    private var startedAt: Long = 0

    private var pausedAt: Long = 0

    val state: StateFlow<PlaybackState> = mutableState.asStateFlow()

    /** Starts playing [schedule]. Returns `false` when a playback is already running. */
    fun start(schedule: MusicSchedule): Boolean {
        if (job?.isActive == true) return false

        val actions: List<PlaybackAction> = timelineOf(schedule)
        paused.value = false
        mutableState.value = PlaybackState(playing = true)
        job = scope.launch { run(actions) }

        return true
    }

    /** Suspends the playback at its current position and releases the held keys. */
    fun pause() {
        if (job?.isActive != true || paused.value) return

        pausedAt = System.nanoTime()
        paused.value = true
        releasePressed()
        mutableState.value = mutableState.value.copy(paused = true)
    }

    /** Continues a paused playback, shifting the timeline by the paused time. */
    fun resume() {
        if (job?.isActive != true || !paused.value) return

        startedAt += System.nanoTime() - pausedAt
        paused.value = false
        mutableState.value = mutableState.value.copy(paused = false)
    }

    fun stop() {
        job?.cancel()
        job = null
        paused.value = false
        mutableState.value = PlaybackState()
    }

    private suspend fun run(actions: List<PlaybackAction>) {
        try {
            delay(leadIn)
            startedAt = System.nanoTime()
            actions.forEach { action -> perform(action) }
        } finally {
            releasePressed()
            mutableState.value = PlaybackState()
        }
    }

    private suspend fun perform(action: PlaybackAction) {
        await(action.begin)

        when (action.kind) {
            ActionKind.ReleaseKey -> {
                sender.releaseKey(action.stroke)
                pressedKeys -= action.stroke
            }

            ActionKind.ReleaseModifiers -> sender.releaseModifiers(action.stroke)

            ActionKind.Press -> {
                sender.press(action.stroke)
                pressedKeys += action.stroke
                mutableState.value = PlaybackState(playing = true, index = action.index, stroke = action.stroke)
            }
        }
    }

    private fun releasePressed() {
        pressedKeys.toList().forEach { stroke -> runCatching { sender.release(stroke) } }
        pressedKeys.clear()
    }

    /**
     * Waits until [offset] has elapsed since the playback started.
     *
     * The wait is suspended while paused, and [resume] moves [startedAt] forward
     * by the paused time, so the remaining actions keep their relative timing.
     */
    private suspend fun await(offset: Duration) {
        while (true) {
            paused.first { isPaused -> !isPaused }

            val target: Long = startedAt + offset.inWholeNanoseconds
            val remaining: Duration = (target - System.nanoTime()).nanoseconds

            if (remaining <= Duration.ZERO) return

            if (remaining <= SLEEP_THRESHOLD) {
                while (System.nanoTime() < target) Thread.onSpinWait()
                return
            }

            delay(minOf(remaining - SLEEP_THRESHOLD, PAUSE_POLL))
        }
    }

    /**
     * Interleaves presses, key releases and modifier releases into one timeline.
     *
     * A release never overlaps the next press of the same key, so repeated notes
     * on one key are cleanly retriggered. Modifiers are lifted before the next
     * note that does not share them, so a plain key can never inherit a chord.
     */
    private fun timelineOf(schedule: MusicSchedule): List<PlaybackAction> {
        val notes: List<ScheduledNote> = schedule.notes
        val nextBegin: MutableMap<KeyStroke, Duration> = mutableMapOf()
        val actions: MutableList<PlaybackAction> = mutableListOf()

        for (index in notes.indices.reversed()) {
            val note: ScheduledNote = notes[index]
            val next: Duration = nextBegin[note.stroke] ?: Duration.INFINITE
            val keyUp: Duration = minOf(note.begin + hold, next)

            actions += PlaybackAction(keyUp, note.stroke, ActionKind.ReleaseKey, index)
            actions += PlaybackAction(note.begin, note.stroke, ActionKind.Press, index)

            if (note.stroke.modifiers.isNotEmpty()) {
                actions += PlaybackAction(modifierUp(notes, index), note.stroke, ActionKind.ReleaseModifiers, index)
            }

            nextBegin[note.stroke] = note.begin
        }

        return actions.sortedWith(ACTION_ORDER)
    }

    /**
     * Returns the moment the modifiers of the note at [index] have to be lifted.
     *
     * A chord only needs its modifiers while its own key-down is delivered, so
     * they are released before the next note that does not share them; holding
     * them longer would make that note look like another chord.
     */
    private fun modifierUp(notes: List<ScheduledNote>, index: Int): Duration {
        val note: ScheduledNote = notes[index]
        val limit: Duration = note.begin + hold
        val conflict: ScheduledNote? = notes.subList(index + 1, notes.size)
            .firstOrNull { other ->
                other.begin > note.begin && !other.stroke.modifiers.containsAll(note.stroke.modifiers)
            }

        return conflict?.let { next -> minOf(limit, next.begin) } ?: limit
    }

    private companion object {
        val DEFAULT_HOLD: Duration = 50.milliseconds
        val DEFAULT_LEAD_IN: Duration = 3.seconds
        val SLEEP_THRESHOLD: Duration = 2.milliseconds
        val PAUSE_POLL: Duration = 50.milliseconds
        val ACTION_ORDER: Comparator<PlaybackAction> = compareBy(
            { action -> action.begin },
            { action -> action.kind.ordinal },
            { action -> action.stroke.modifiers.size }
        )
    }

}

/** Order of the actions scheduled at the same instant: releases first. */
private enum class ActionKind { ReleaseKey, ReleaseModifiers, Press }

private data class PlaybackAction(
    val begin: Duration,
    val stroke: KeyStroke,
    val kind: ActionKind,
    val index: Int
)
