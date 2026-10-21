package dev.xueria.nte.tools.language

import dev.xueria.nte.tools.navigation.NavigationItem

typealias TranslateKey = String

fun translate(translateKey: TranslateKey, defaultText: String? = null): String {
    val fallback = translateKey.split(".", "_")
        .joinToString("") { part ->
            part.replaceFirstChar { it.uppercase() }
        }
    return LanguageManager.translate(translateKey) ?: defaultText ?: fallback
}

object Translate {

    object Application {

        fun name() = translate("app.name", "Application")

    }

    object Navigation {

        fun label(item: NavigationItem): String {
            return translate("navigation.${item.id}.label", item.label)
        }

        fun title(item: NavigationItem): String {
            return translate("navigation.${item.id}.title", item.title)
        }

        fun description(item: NavigationItem): String {
            return translate("navigation.${item.id}.description", item.description)
        }


    }

}