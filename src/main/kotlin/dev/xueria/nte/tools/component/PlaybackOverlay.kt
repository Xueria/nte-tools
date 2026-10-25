package dev.xueria.nte.tools.component

import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Window
import androidx.compose.ui.window.WindowPosition
import androidx.compose.ui.window.rememberWindowState
import dev.xueria.nte.tools.language.translate
import dev.xueria.nte.tools.music.KeyStroke
import dev.xueria.nte.tools.music.play.PlaybackState
import dev.xueria.nte.tools.theme.Theme
import java.awt.MouseInfo
import java.awt.Point
import java.awt.Toolkit
import java.awt.event.KeyEvent

private val OverlayWidth: Dp = 420.dp
private val OverlayHeight: Dp = 96.dp
private val OverlayMargin: Dp = 10.dp
private val OverlayShape: RoundedCornerShape = RoundedCornerShape(18.dp)

/** Fraction of the screen height where the overlay starts. */
private const val OverlayTopRatio: Float = 0.12f

/**
 * Always-on-top window shown while a sheet is playing.
 *
 * It never takes the keyboard focus, otherwise the forwarded key presses would
 * land in the overlay instead of the game window. It can be dragged anywhere by
 * its card; the buttons keep working because a click never passes the touch slop.
 */
@Composable
fun PlaybackOverlay(
    title: String,
    state: PlaybackState,
    onTogglePause: () -> Unit,
    onStop: () -> Unit
) {
    val density = LocalDensity.current

    val position: WindowPosition = remember(density) {
        val screen = Toolkit.getDefaultToolkit().screenSize
        val width: Int = with(density) { OverlayWidth.roundToPx() }

        WindowPosition(
            x = with(density) { ((screen.width - width) / 2).toDp() },
            y = with(density) { (screen.height * OverlayTopRatio).toInt().toDp() }
        )
    }

    Window(
        onCloseRequest = onStop,
        state = rememberWindowState(
            position = position,
            size = DpSize(OverlayWidth, OverlayHeight)
        ),
        title = title,
        resizable = false,
        undecorated = true,
        transparent = true,
        focusable = false,
        alwaysOnTop = true
    ) {
        Theme {
            Surface(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(OverlayMargin)
                    .pointerInput(Unit) {
                        var grabOffset: Point? = null

                        detectDragGestures(
                            onDragStart = {
                                grabOffset = MouseInfo.getPointerInfo()?.location?.let { pointer ->
                                    Point(pointer.x - window.x, pointer.y - window.y)
                                }
                            },
                            onDragEnd = { grabOffset = null },
                            onDragCancel = { grabOffset = null },
                            onDrag = { change, _ ->
                                change.consume()

                                val offset: Point? = grabOffset
                                val pointer: Point? = MouseInfo.getPointerInfo()?.location

                                if (offset != null && pointer != null) {
                                    window.setLocation(pointer.x - offset.x, pointer.y - offset.y)
                                }
                            }
                        )
                    },
                shape = OverlayShape,
                color = MaterialTheme.colorScheme.surfaceContainerHigh,
                shadowElevation = 12.dp
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(horizontal = 20.dp, vertical = 12.dp),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(16.dp)
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = title,
                            style = MaterialTheme.typography.titleMedium,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis
                        )

                        Text(
                            text = translate("music_player.current_key", "Current key"),
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }

                    Text(
                        text = state.stroke?.readable()
                            ?: translate("music_player.starting", "Starting"),
                        style = MaterialTheme.typography.headlineSmall
                    )

                    Button(onClick = onTogglePause) {
                        Text(
                            text = if (state.paused) {
                                translate("music_player.resume", "Resume")
                            } else {
                                translate("music_player.pause", "Pause")
                            }
                        )
                    }

                    TextButton(onClick = onStop) {
                        Text(translate("music_player.stop", "Stop"))
                    }
                }
            }
        }
    }
}

private fun KeyStroke.readable(): String =
    (modifiers.map { modifier -> KeyEvent.getKeyText(modifier.keyCode) } + KeyEvent.getKeyText(keyCode))
        .joinToString("+")
