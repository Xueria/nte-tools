package dev.xueria.nte.tools.music.play

import dev.xueria.nte.tools.music.KeyStroke
import java.awt.Robot

/** Sends real key presses to the operating system. */
interface KeySender {

    fun press(stroke: KeyStroke)

    /** Releases the main key and every modifier of [stroke]. */
    fun release(stroke: KeyStroke)

    /** Releases only the main key, keeping the modifiers of a chord held. */
    fun releaseKey(stroke: KeyStroke)

    /** Releases only the modifiers, so the next note cannot inherit them. */
    fun releaseModifiers(stroke: KeyStroke)

}

/** A [KeySender] backed by [Robot]. */
class RobotKeySender(private val robot: Robot) : KeySender {

    // Pause is called from the UI thread while the playback coroutine keeps
    // sending keys, and Robot is not thread-safe.
    @Synchronized
    override fun press(stroke: KeyStroke) {
        stroke.modifiers.forEach { modifier -> robot.keyPress(modifier.keyCode) }
        robot.keyPress(stroke.keyCode)
    }

    @Synchronized
    override fun release(stroke: KeyStroke) {
        releaseKey(stroke)
        releaseModifiers(stroke)
    }

    @Synchronized
    override fun releaseKey(stroke: KeyStroke) {
        robot.keyRelease(stroke.keyCode)
    }

    @Synchronized
    override fun releaseModifiers(stroke: KeyStroke) {
        stroke.modifiers.toList().asReversed().forEach { modifier -> robot.keyRelease(modifier.keyCode) }
    }

    companion object {

        /** Returns `null` when the platform cannot create a robot, for example in a headless session. */
        fun createOrNull(): RobotKeySender? = runCatching { RobotKeySender(Robot()) }.getOrNull()

    }

}
