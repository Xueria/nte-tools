package dev.xueria.nte.tools

import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import org.jetbrains.compose.resources.painterResource

fun main() {
    configureRenderBackend()
    application {
        Window(
            onCloseRequest = ::exitApplication,
            title = "nte-music",
            icon = painterResource(Res.drawable.windows)
        ) {
            ApplicationContent()
        }
    }
}

private fun configureRenderBackend() {
    val osName = System.getProperty("os.name").orEmpty()
    if (osName.contains("windows", true)) {
        System.setProperty("skiko.renderApi", "OPENGL")
    }
}