package dev.xueria.nte.tools

import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import dev.xueria.nte.tools.language.LanguageManager
import dev.xueria.nte.tools.platform.RunningPath
import dev.xueria.nte.tools.shell.Destination
import dev.xueria.nte.tools.shell.DestinationPane
import dev.xueria.nte.tools.shell.MainDestinations
import dev.xueria.nte.tools.shell.Sidebar
import dev.xueria.nte.tools.shell.SidebarItem
import dev.xueria.nte.tools.theme.Theme

@Composable
@Preview
fun ApplicationContent() {
    Theme {
        MainScreen()
    }
}

@Composable
fun MainScreen() {
    LaunchedEffect(Unit) {
        LanguageManager.loadAll()
    }

    var selectedDestination: Destination by remember {
        mutableStateOf(Destination.MusicPlayer)
    }

    Row(
        modifier = Modifier.fillMaxSize()
    ) {
        Sidebar(
            destinations = MainDestinations,
            selectedDestination = selectedDestination,
            onDestinationSelected = { destination -> selectedDestination = destination },
            showBottomDivider = true
        ) {
            SidebarItem(
                destination = Destination.Setting,
                selected = selectedDestination == Destination.Setting,
                onClick = {
                    selectedDestination = Destination.Setting
                }
            )
        }

        DestinationPane(destination = selectedDestination, modifier = Modifier.weight(1f))
    }
}
