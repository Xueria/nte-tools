package dev.xueria.nte.tools.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import dev.xueria.nte.tools.language.LanguageManager
import dev.xueria.nte.tools.language.Translate
import dev.xueria.nte.tools.language.translate
import dev.xueria.nte.tools.shell.Destination

@Composable
fun SettingScreen(modifier: Modifier = Modifier) {
    Column(
        modifier = modifier.padding(32.dp)
    ) {
        Text(
            text = Translate.Navigation.title(Destination.Setting),
            style = MaterialTheme.typography.headlineMedium
        )

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = Translate.Navigation.description(Destination.Setting),
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        Spacer(modifier = Modifier.height(24.dp))

        Text(
            text = translate("settings.language", "Language"),
            style = MaterialTheme.typography.titleMedium
        )

        Spacer(modifier = Modifier.height(8.dp))

        Row(
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            TextButton(
                onClick = { LanguageManager.changeLanguage("zh_cn") }
            ) {
                Text("中文")
            }

            TextButton(
                onClick = { LanguageManager.changeLanguage("en_us") }
            ) {
                Text("English")
            }
        }

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = LanguageManager.language?.name.orEmpty(),
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}