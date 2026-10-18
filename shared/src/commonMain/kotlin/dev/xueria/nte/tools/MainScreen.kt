package dev.xueria.nte.tools

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.sp
import dev.xueria.nte.tools.language.LanguageManager
import dev.xueria.nte.tools.language.translate
import dev.xueria.nte.tools.navigation.Navigation
import dev.xueria.nte.tools.theme.Colors
import dev.xueria.nte.tools.theme.Dimens

@Composable
@Preview
fun MainScreen() {
    LaunchedEffect(Unit) {
        LanguageManager.loadAll()
    }

    Row(
        modifier = Modifier.fillMaxSize()
    ) {
        Box(
            modifier = Modifier
                .width(Dimens.SidebarWidth)
                .fillMaxHeight()
                .background(Colors.SidebarBackground)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(
                        horizontal = Dimens.SidebarHorizontalPadding,
                        vertical = Dimens.SidebarVerticalPadding
                    ),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Navigation.items.forEach { item ->
                    Text(
                        text = translate(item.labelTranslateKey),
                        modifier = Modifier.fillMaxWidth(),
                        fontSize = 12.sp,
                        textAlign = TextAlign.Center
                    )
                }
            }
        }

        Box(
            modifier = Modifier.weight(1f)
                .fillMaxHeight()
                .background(Colors.ContentBackground)
        )
    }
}
