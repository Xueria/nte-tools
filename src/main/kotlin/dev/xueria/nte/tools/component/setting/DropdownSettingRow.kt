package dev.xueria.nte.tools.component.setting

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import dev.xueria.nte.tools.Res
import dev.xueria.nte.tools.ic_check
import dev.xueria.nte.tools.ic_expand
import dev.xueria.nte.tools.model.Selection
import org.jetbrains.compose.resources.painterResource

@Composable
fun <T : Selection> DropdownSettingRow(
    label: String,
    options: List<T>,
    selected: T,
    onSelect: (T) -> Unit,
    modifier: Modifier = Modifier
) {
    SettingRow(label = label, modifier = modifier) {
        DropdownField(
            options = options,
            selected = selected,
            onSelect = onSelect
        )
    }
}

private val DropdownFieldShape: RoundedCornerShape = androidx.compose.foundation.shape.RoundedCornerShape(8.dp)
private val DropdownFieldPadding: Dp = 12.dp
private val DropdownFieldMinWidth: Dp = 160.dp
private val DropdownIconSize: Dp = 16.dp
private val DropdownMenuMaxHeight: Dp = 320.dp

@Composable
fun <T : Selection> DropdownField(
    options: List<T>,
    selected: T,
    onSelect: (T) -> Unit,
    modifier: Modifier = Modifier
) {
    var expanded: Boolean by remember { mutableStateOf(false) }

    Box(modifier = modifier) {
        DropdownFieldSurface(
            text = selected.label,
            expanded = expanded,
            onClick = { expanded = !expanded }
        )

        DropdownMenu(
            expanded = expanded,
            onDismissRequest = { expanded = false },
            modifier = Modifier
                .widthIn(min = DropdownFieldMinWidth)
                .heightIn(max = DropdownMenuMaxHeight)
        ) {
            options.forEach { option ->
                DropdownOption(
                    label = option.label,
                    selected = option == selected,
                    onClick = {
                        expanded = false
                        onSelect(option)
                    }
                )
            }
        }
    }
}

@Composable
private fun DropdownFieldSurface(
    text: String,
    expanded: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val borderColor: Color = if (expanded) {
        MaterialTheme.colorScheme.primary
    } else {
        MaterialTheme.colorScheme.outline
    }

    Row(
        modifier = modifier
            .widthIn(min = DropdownFieldMinWidth)
            .clip(DropdownFieldShape)
            .border(1.dp, borderColor, DropdownFieldShape)
            .background(MaterialTheme.colorScheme.surface)
            .clickable(onClick = onClick)
            .padding(DropdownFieldPadding),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(
            text = text,
            style = MaterialTheme.typography.bodyMedium,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.weight(1f, fill = false)
        )

        DropdownChevron(
            modifier = Modifier.rotate(if (expanded) 180f else 0f)
        )
    }
}

@Composable
private fun DropdownOption(
    label: String,
    selected: Boolean,
    onClick: () -> Unit
) {
    DropdownMenuItem(
        text = {
            Text(
                text = label,
                style = MaterialTheme.typography.bodyMedium,
                color = if (selected) {
                    MaterialTheme.colorScheme.primary
                } else {
                    MaterialTheme.colorScheme.onSurface
                },
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )
        },
        onClick = onClick,
        trailingIcon = if (selected) {
            { DropdownSelectionCheck() }
        } else {
            null
        }
    )
}

@Composable
private fun DropdownSelectionCheck() {
    Icon(
        painter = painterResource(Res.drawable.ic_check),
        contentDescription = null,
        tint = MaterialTheme.colorScheme.primary,
        modifier = Modifier.size(DropdownIconSize)
    )
}

@Composable
private fun DropdownChevron(modifier: Modifier = Modifier) {
    Icon(
        painter = painterResource(Res.drawable.ic_expand),
        contentDescription = null,
        tint = MaterialTheme.colorScheme.onSurfaceVariant,
        modifier = modifier.size(DropdownIconSize)
    )
}