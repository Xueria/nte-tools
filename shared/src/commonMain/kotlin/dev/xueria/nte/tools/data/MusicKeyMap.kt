package dev.xueria.nte.tools.data

import kotlinx.serialization.Serializable

@Serializable
data class MusicKeyMap(
    val type: String,
    val keymap: Map<String, String>
)