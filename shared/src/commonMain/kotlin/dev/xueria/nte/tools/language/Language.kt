package dev.xueria.nte.tools.language

import kotlinx.serialization.Serializable

@Serializable
data class Language(
    val locale: String,
    val name: String,
    val messages: Map<String, String>
)