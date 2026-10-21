package dev.xueria.nte.tools.shell

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import dev.xueria.nte.tools.theme.Dimens

@Composable
fun Sidebar(
    destinations: List<Destination>,
    selectedDestination: Destination,
    onDestinationSelected: (Destination) -> Unit,
    modifier: Modifier = Modifier,
    showBottomDivider: Boolean = false,
    bottomContent: @Composable ColumnScope.() -> Unit = {}
) {
    Box(
        modifier = modifier
            .width(Dimens.Sidebar.Width)
            .fillMaxHeight()
            .background(MaterialTheme.colorScheme.surfaceContainer)
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
                    .padding(
                        horizontal = Dimens.Sidebar.HorizontalPadding,
                        vertical = Dimens.Sidebar.VerticalPadding
                    )
                    .verticalScroll(rememberScrollState()),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(Dimens.NavigationItemSpacing)
            ) {
                destinations.forEach { destination ->
                    SidebarItem(
                        destination = destination,
                        selected = destination == selectedDestination,
                        onClick = { onDestinationSelected(destination) }
                    )
                }
            }

            if (showBottomDivider) {
                HorizontalDivider(
                    modifier = Modifier.fillMaxWidth(),
                    color = MaterialTheme.colorScheme.outlineVariant
                )
            }

            Column(
                modifier = Modifier.fillMaxWidth()
                    .padding(
                        horizontal = Dimens.Sidebar.HorizontalPadding,
                        vertical = Dimens.Sidebar.BottomVerticalPadding
                    ),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(Dimens.NavigationItemSpacing)
            ) {
                bottomContent()
            }
        }


    }
}