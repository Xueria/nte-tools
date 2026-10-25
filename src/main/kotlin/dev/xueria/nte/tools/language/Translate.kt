package dev.xueria.nte.tools.language

import dev.xueria.nte.tools.shell.Destination

typealias TranslateKey = String

fun translate(translateKey: TranslateKey, default: String? = null): String {
    return LanguageManager.language.translate(key = translateKey, default = default)
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