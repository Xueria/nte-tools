package dev.xueria.nte.tools.language

import kotlinx.serialization.Serializable

data class TranslateKey(val key: String) {

    companion object {

        private const val DOT = "."

    }

    infix fun dot(subkey: String): TranslateKey {
        return TranslateKey(this@TranslateKey.key + DOT + subkey)
    }

    fun toReadableName(): String = key
        .split(DOT, "_")
        .joinToString("") { part ->
            part.replaceFirstChar { it.uppercase() }
        }

}

@Serializable
data class Language(
    val locale: String,
    val name: String,
    val messages: Map<String, String>
)