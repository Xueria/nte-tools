package dev.xueria.nte.tools.screen

import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.painter.Painter

abstract class Destination(
    val label: String,
    val icon: Painter? = null
) {

    @Composable
    abstract fun RenderScreen()

}
