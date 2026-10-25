package dev.xueria.nte.tools.language

import dev.xueria.nte.tools.model.Selection
import kotlinx.serialization.Serializable

sealed interface Language : Selection {

    fun locale(): String

    fun translate(key: TranslateKey, default: String?): String

    data object None : Language {

        override val label: String
            get() = "None"

        override fun locale(): String = "none"

        override fun translate(key: TranslateKey, default: String?): String =
            default ?: key.fallback
    }

    @Serializable
    data class Default(
        val locale: String,
        val name: String,
        val messages: Map<String, String>
    ) : Language {

        override val label: String
            get() = name

        override fun locale(): String = locale

        override fun translate(key: TranslateKey, default: String?): String =
            messages[key] ?: default ?: key.fallback

    }

}

private val TranslateKey.fallback
    get() = split(".", "_", "-")
        .joinToString("") { part ->
            part.replaceFirstChar { it.uppercase() }
        }