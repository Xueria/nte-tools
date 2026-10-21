package dev.xueria.nte.tools.screen

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import dev.xueria.nte.tools.language.Translate
import dev.xueria.nte.tools.shell.Destination

@Composable
fun MusicPlayerScreen(modifier: Modifier = Modifier) {
    Column(
        modifier = modifier.padding(32.dp)
    ) {
        Text(
            text = Translate.Navigation.title(Destination.MusicPlayer),
            style = MaterialTheme.typography.headlineMedium
        )

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = Translate.Navigation.description(Destination.MusicPlayer),
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}