package dev.xueria.nte.tools.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import dev.xueria.nte.tools.theme.color.pink.PinkDarkScheme
import dev.xueria.nte.tools.theme.color.pink.PinkLightScheme

@Composable
fun Theme(
    isDarkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit
) {

    val colorScheme = if (isDarkTheme) PinkDarkScheme else PinkLightScheme

    MaterialTheme(
        colorScheme = colorScheme,
        content = content
    )
}