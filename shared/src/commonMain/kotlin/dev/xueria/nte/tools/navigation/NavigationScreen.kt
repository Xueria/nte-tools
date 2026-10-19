package dev.xueria.nte.tools.navigation

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.selection.selectable
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.semantics.Role
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp
import dev.xueria.nte.tools.theme.Dimens

@Composable
fun NavigationSidebarItem(
    item: NavigationItem,
    selected: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val rowColor: Color =
        if (selected) MaterialTheme.colorScheme.secondaryContainer else Color.Transparent
    val iconColor: Color =
        if (selected) MaterialTheme.colorScheme.onSecondaryContainer else MaterialTheme.colorScheme.onSurface
    val labelColor: Color =
        if (selected) MaterialTheme.colorScheme.onSecondaryContainer else MaterialTheme.colorScheme.onSurfaceVariant

    Row(
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(Dimens.NavigationItemRadius))
            .background(rowColor)
            .selectable(
                selected = selected,
                onClick = onClick,
                role = Role.Tab
            )
            .padding(
                horizontal = Dimens.NavigationItemHorizontalPadding,
                vertical = Dimens.NavigationItemVerticalPadding
            ),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(Dimens.NavigationLabelSpacing)
    ) {
        Box(
            modifier = Modifier.size(Dimens.NavigationIconTileSize),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = item.icon.placeholderGlyph(),
                color = iconColor,
                fontSize = 12.sp,
                fontWeight = FontWeight.SemiBold,
            )
        }

        Text(
            text = item.label,
            color = labelColor,
            fontSize = 12.sp,
            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
            maxLines = 1,
        )
    }
}

fun NavigationIcon.placeholderGlyph(): String = when (this) {
    NavigationIcon.MusicPlayer -> "音乐"
}