package dev.xueria.nte.tools

import androidx.compose.ui.window.Window
import androidx.compose.ui.window.application
import org.jetbrains.compose.resources.painterResource

fun main() = application {
    Window(
        onCloseRequest = ::exitApplication,
        title = "nte-music",
        icon = painterResource(Res.drawable.windows)
    ) {
        MainScreen()
    }
}