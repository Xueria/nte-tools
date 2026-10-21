package dev.xueria.nte.tools.shell

import androidx.compose.material3.Icon
import androidx.compose.material3.LocalContentColor
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.painter.Painter
import dev.xueria.nte.tools.Res
import dev.xueria.nte.tools.ic_music_player
import dev.xueria.nte.tools.ic_setting
import org.jetbrains.compose.resources.painterResource

@Composable
fun IconSelector(
    destination: Destination,
    modifier: Modifier = Modifier,
    tint: Color = LocalContentColor.current
) {
    val painter: Painter = when (destination) {
        Destination.MusicPlayer -> painterResource(Res.drawable.ic_music_player)
        Destination.Setting -> painterResource(Res.drawable.ic_setting)
    }

    Icon(
        painter = painter,
        contentDescription = null,
        tint = tint,
        modifier = modifier
    )
}