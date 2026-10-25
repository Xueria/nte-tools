package dev.xueria.nte.tools.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import dev.xueria.nte.tools.component.setting.DropdownSettingRow
import dev.xueria.nte.tools.language.Language
import dev.xueria.nte.tools.language.LanguageManager
import dev.xueria.nte.tools.language.translate

private val SettingItemSpacing: Dp = 4.dp

@Composable
fun SettingScreen(modifier: Modifier = Modifier) {
    Column(
        modifier = modifier.padding(horizontal = 32.dp, vertical = 24.dp),
        verticalArrangement = Arrangement.spacedBy(SettingItemSpacing)
    ) {
        LanguageSetting(
            modifier = Modifier
                .fillMaxWidth()
        )
    }
}

@Composable
private fun LanguageSetting(modifier: Modifier = Modifier) {
    val languages: List<Language> = LanguageManager.languages.values.toList()
    val selected: Language = LanguageManager.language

    if (languages.isEmpty()) {
        return
    }

    DropdownSettingRow(
        label = translate("settings.language", "Language"),
        options = languages,
        selected = selected,
        onSelect = { language -> LanguageManager.changeLanguage(language.locale()) },
        modifier = modifier
    )
}
