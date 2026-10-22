package dev.xueria.nte.tools.shell

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import dev.xueria.nte.tools.screen.MusicPlayerScreen
import dev.xueria.nte.tools.screen.SettingScreen

@Composable
fun DestinationPane(
    destination: Destination,
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier
            .fillMaxHeight()
            .background(MaterialTheme.colorScheme.background)
    ) {
        when (destination) {
            Destination.MusicPlayer -> MusicPlayerScreen(Modifier.fillMaxSize())
            Destination.Setting -> SettingScreen(Modifier.fillMaxSize())
        }
    }

}