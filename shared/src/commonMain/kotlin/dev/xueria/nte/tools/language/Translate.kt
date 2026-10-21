package dev.xueria.nte.tools.language

import dev.xueria.nte.tools.shell.Destination

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

        fun label(destination: Destination): String {
            return translate("navigation.${destination.id}.label", destination.label)
        }

        fun title(destination: Destination): String {
            return translate("navigation.${destination.id}.title", destination.title)
        }

        fun description(destination: Destination): String {
            return translate("navigation.${destination.id}.description", destination.description)
        }


    }

}