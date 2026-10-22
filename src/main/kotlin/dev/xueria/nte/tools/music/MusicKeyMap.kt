package dev.xueria.nte.tools.music

import kotlinx.serialization.Serializable

@Serializable
data class MusicKeyMap(
    val type: String,
    val keymap: Map<String, String>
)