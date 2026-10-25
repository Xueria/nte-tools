package dev.xueria.nte.tools.screen

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import dev.xueria.nte.tools.component.PlaybackOverlay
import dev.xueria.nte.tools.component.setting.DropdownSettingRow
import dev.xueria.nte.tools.language.translate
import dev.xueria.nte.tools.music.BakedKeyMapper
import dev.xueria.nte.tools.music.ExKeyMapper
import dev.xueria.nte.tools.music.KeyMapper
import dev.xueria.nte.tools.music.Music
import dev.xueria.nte.tools.music.MusicKeyMaps
import dev.xueria.nte.tools.music.MusicSheet
import dev.xueria.nte.tools.music.play.AutoPlayer
import dev.xueria.nte.tools.music.play.PlaybackPlan
import dev.xueria.nte.tools.music.play.PlaybackState
import dev.xueria.nte.tools.music.play.RobotKeySender
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import java.util.Locale
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

private val ScreenPadding: Dp = 32.dp
private val ScreenSpacing: Dp = 24.dp
private val ContentSpacing: Dp = 12.dp
private val SongListWidth: Dp = 220.dp
private val SongRowSpacing: Dp = 6.dp
private val SongRowShape: RoundedCornerShape = RoundedCornerShape(10.dp)
private val RefreshInterval: Duration = 2.seconds

@Composable
fun MusicPlayerScreen(modifier: Modifier = Modifier) {
    LaunchedEffect(Unit) {
        while (isActive) {
            delay(RefreshInterval)
            Music.refresh()
            MusicKeyMaps.refresh()
        }
    }

    val sheets: Map<String, MusicSheet> = Music.sheets
    val keyMaps: Map<String, ExKeyMapper> = MusicKeyMaps.keyMaps

    if (sheets.isEmpty()) {
        EmptyMessage(modifier = modifier)
        return
    }

    val scope: CoroutineScope = rememberCoroutineScope()
    val player: AutoPlayer? = remember(scope) {
        RobotKeySender.createOrNull()?.let { sender -> AutoPlayer(sender, scope = scope) }
    }

    var selectedSheetId: String? by remember { mutableStateOf(null) }
    var selectedKeyMap: String? by remember { mutableStateOf(null) }

    val sheetId: String = selectedSheetId?.takeIf { id -> id in sheets } ?: sheets.keys.first()
    val sheet: MusicSheet = sheets.getValue(sheetId)
    val exKeyMap: ExKeyMapper? = selectedKeyMap?.let { name -> keyMaps[name] } ?: keyMaps.values.firstOrNull()
    val keyMapper: KeyMapper? = remember(exKeyMap) { exKeyMap?.let { mapper -> BakedKeyMapper(mapper) } }
    val plan: PlaybackPlan? = remember(sheet, keyMapper) {
        keyMapper?.let { mapper -> runCatching { PlaybackPlan(sheet, mapper) }.getOrNull() }
    }
    val playback: PlaybackState = player?.state?.collectAsState()?.value ?: PlaybackState()

    LaunchedEffect(sheetId, exKeyMap) {
        player?.stop()
    }

    Row(
        modifier = modifier.padding(horizontal = ScreenPadding, vertical = 24.dp),
        horizontalArrangement = Arrangement.spacedBy(ScreenSpacing)
    ) {
        SongList(
            sheets = sheets,
            selectedId = sheetId,
            onSelect = { id -> selectedSheetId = id },
            modifier = Modifier
                .width(SongListWidth)
                .fillMaxHeight()
        )

        Box(
            modifier = Modifier
                .fillMaxHeight()
                .width(1.dp)
                .background(MaterialTheme.colorScheme.outlineVariant)
        )

        SongDetail(
            sheet = sheet,
            plan = plan,
            keyMaps = keyMaps.values.toList(),
            selectedKeyMap = exKeyMap,
            onKeyMapSelect = { mapper -> selectedKeyMap = mapper.name },
            player = player,
            state = playback,
            modifier = Modifier.weight(1f)
        )
    }

    if (playback.playing) {
        PlaybackOverlay(
            title = sheetId,
            state = playback,
            onTogglePause = {
                if (playback.paused) player?.resume() else player?.pause()
            },
            onStop = { player?.stop() }
        )
    }
}

@Composable
private fun EmptyMessage(modifier: Modifier = Modifier) {
    Box(
        modifier = modifier
            .fillMaxSize()
            .padding(ScreenPadding),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = translate("music_player.empty", "No sheets loaded."),
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

@Composable
private fun SongList(
    sheets: Map<String, MusicSheet>,
    selectedId: String,
    onSelect: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    val rows: List<SongItem> = remember(sheets) {
        sheets.entries.map { entry ->
            SongItem(id = entry.key, duration = entry.value.sheetDuration.readable())
        }
    }

    LazyColumn(
        modifier = modifier,
        verticalArrangement = Arrangement.spacedBy(SongRowSpacing)
    ) {
        items(items = rows, key = { row -> row.id }) { row ->
            SongRow(
                title = row.id,
                subtitle = row.duration,
                selected = row.id == selectedId,
                onClick = { onSelect(row.id) }
            )
        }
    }
}

/** One row of the song list: the file name plus its length. */
private data class SongItem(
    val id: String,
    val duration: String
)

@Composable
private fun SongRow(
    title: String,
    subtitle: String,
    selected: Boolean,
    onClick: () -> Unit
) {
    val background = if (selected) {
        MaterialTheme.colorScheme.secondaryContainer
    } else {
        MaterialTheme.colorScheme.surfaceVariant
    }

    Column(
        modifier = Modifier
            .fillMaxWidth()
            .clip(SongRowShape)
            .background(background)
            .clickable(onClick = onClick)
            .padding(horizontal = 12.dp, vertical = 10.dp)
    ) {
        Text(
            text = title,
            style = MaterialTheme.typography.titleSmall,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis
        )

        Text(
            text = subtitle,
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis
        )
    }
}

@Composable
private fun SongDetail(
    sheet: MusicSheet,
    plan: PlaybackPlan?,
    keyMaps: List<ExKeyMapper>,
    selectedKeyMap: ExKeyMapper?,
    onKeyMapSelect: (ExKeyMapper) -> Unit,
    player: AutoPlayer?,
    state: PlaybackState,
    modifier: Modifier = Modifier
) {
    val removed: Int = plan?.removedNotes ?: 0

    Column(
        modifier = modifier,
        verticalArrangement = Arrangement.spacedBy(ContentSpacing)
    ) {
        Text(
            text = sheet.name,
            style = MaterialTheme.typography.headlineSmall
        )

        HorizontalDivider()

        InfoRow(translate("music_player.bpm", "BPM"), sheet.bpm.readableBpm())
        InfoRow(
            label = translate("music_player.time_signature", "Time signature"),
            value = "${sheet.timeSignature.top}/${sheet.timeSignature.bottom}"
        )
        InfoRow(
            label = translate("music_player.notes", "Notes"),
            value = "${plan?.playableNotes ?: 0} / ${plan?.totalNotes ?: 0}"
        )
        InfoRow(translate("music_player.duration", "Duration"), plan?.duration?.readable() ?: "-")

        if (removed > 0) {
            Text(
                text = "${translate("music_player.unplayable", "Unplayable notes")}: $removed",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.error
            )
        }

        if (selectedKeyMap != null && keyMaps.isNotEmpty()) {
            DropdownSettingRow(
                label = translate("music_player.keymap", "Keymap"),
                options = keyMaps,
                selected = selectedKeyMap,
                onSelect = onKeyMapSelect
            )
        }

        PlaybackControls(plan = plan, player = player, state = state)
    }
}

@Composable
private fun InfoRow(label: String, value: String) {
    Row(
        modifier = Modifier.fillMaxWidth(),
        horizontalArrangement = Arrangement.SpaceBetween
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        Text(
            text = value,
            style = MaterialTheme.typography.bodyMedium
        )
    }
}

@Composable
private fun PlaybackControls(plan: PlaybackPlan?, player: AutoPlayer?, state: PlaybackState) {
    val total: Int = plan?.schedule?.notes?.size ?: 0
    val progress: String = if (state.index >= 0) "${state.index + 1} / $total" else "-"

    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        Button(
            onClick = {
                if (state.playing) {
                    player?.stop()
                } else {
                    plan?.let { ready -> player?.start(ready.schedule) }
                }
            },
            enabled = (player != null && total > 0) || state.playing
        ) {
            Text(
                text = if (state.playing) {
                    translate("music_player.stop", "Stop")
                } else {
                    translate("music_player.play", "Play")
                }
            )
        }

        Text(
            text = progress,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }

    if (player == null) {
        Text(
            text = translate("music_player.no_keyboard", "Cannot create a key sender."),
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.error
        )
    }

    if (state.playing && state.index < 0) {
        Text(
            text = translate("music_player.lead_in", "Starting in 3 seconds. Switch to the game window."),
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.primary
        )
    }
}

/** Length of the sheet before it is folded into a keymap. */
private val MusicSheet.sheetDuration: Duration
    get() {
        if (bpm <= 0f) return Duration.ZERO

        val secondsPerBeat: Double = SECONDS_PER_MINUTE / bpm
        val beats: Double = tracks.maxOfOrNull { track -> track.notes.sumOf { note -> note.beats.toDouble() } } ?: 0.0

        return (beats * secondsPerBeat).seconds
    }

private const val SECONDS_PER_MINUTE: Double = 60.0

private fun Float.readableBpm(): String =
    String.format(Locale.ROOT, "%.2f", this).trimEnd('0').trimEnd('.')

private fun Duration.readable(): String =
    "%d:%02d".format(inWholeSeconds / 60, inWholeSeconds % 60)