package dev.xueria.nte.tools

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import dev.xueria.nte.tools.language.LanguageManager
import dev.xueria.nte.tools.navigation.MusicPlayer
import dev.xueria.nte.tools.navigation.NavigationItem
import dev.xueria.nte.tools.navigation.NavigationItems
import dev.xueria.nte.tools.navigation.NavigationSidebarItem
import dev.xueria.nte.tools.theme.Dimens
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

    var selectedItemId: String by remember {
        mutableStateOf(MusicPlayer.id)
    }

    val selectedItem: NavigationItem =
        NavigationItems.first { item -> item.id == selectedItemId }

    Row(
        modifier = Modifier.fillMaxSize()
    ) {
        Box(
            modifier = Modifier
                .width(Dimens.SidebarWidth)
                .fillMaxHeight()
                .background(MaterialTheme.colorScheme.surfaceContainer)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(
                        horizontal = Dimens.SidebarHorizontalPadding,
                        vertical = Dimens.SidebarVerticalPadding
                    ),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(Dimens.NavigationItemSpacing)
            ) {
                NavigationItems.forEach { item ->
                    NavigationSidebarItem(
                        item = item,
                        selected = selectedItemId == item.id,
                        onClick = {
                            selectedItemId = item.id
                        }
                    )
                }
            }
        }

        Box(
            modifier = Modifier.weight(1f)
                .fillMaxHeight()
                .background(MaterialTheme.colorScheme.background)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(32.dp)
            ) {
                Text(
                    text = selectedItem.title,
                    maxLines = 1,
                    style = MaterialTheme.typography.headlineMedium
                )
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = selectedItem.description,
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}
