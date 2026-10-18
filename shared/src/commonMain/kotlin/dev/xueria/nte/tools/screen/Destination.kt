package dev.xueria.nte.tools.screen

import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.painter.Painter
import dev.xueria.nte.tools.language.TranslateKey

abstract class Destination(
    val label: String,
    val icon: Painter? = null
) {

    private val translateKey = TranslateKey("navigation").dot(label.lowercase())

    @Composable
    abstract fun RenderScreen()

}
