package dev.xueria.nte.tools

import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import dev.xueria.nte.tools.language.Translate
import org.jetbrains.compose.resources.painterResource

fun main() {
    configureRenderBackend()
    application {
        Window(
            onCloseRequest = ::exitApplication,
            title = Translate.Application.name(),
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