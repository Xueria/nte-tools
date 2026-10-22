package dev.xueria.nte.tools.theme.font

import androidx.compose.runtime.Composable
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import dev.xueria.nte.tools.Res
import dev.xueria.nte.tools.SarasaBold
import dev.xueria.nte.tools.SarasaExtraLight
import dev.xueria.nte.tools.SarasaLight
import dev.xueria.nte.tools.SarasaRegular
import dev.xueria.nte.tools.SarasaSemiBold
import org.jetbrains.compose.resources.Font

@Composable
fun AppFontFamily(): FontFamily {
    return FontFamily(
        Font(Res.font.SarasaRegular, FontWeight.Normal),
        Font(Res.font.SarasaBold, FontWeight.Bold),
        Font(Res.font.SarasaLight, FontWeight.Light),
        Font(Res.font.SarasaExtraLight, FontWeight.ExtraLight),
        Font(Res.font.SarasaSemiBold, FontWeight.SemiBold)
    )
}