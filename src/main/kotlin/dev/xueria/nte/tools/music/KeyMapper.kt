package dev.xueria.nte.tools.music

import dev.xueria.nte.tools.model.Selection
import kotlinx.serialization.Serializable

interface KeyMapper {

    fun keyof(pitch: Pitch): KeyStroke?

}

@Serializable
data class ExKeyMapper(
    val name: String,
    val map: Map<String, String>
) : Selection {

    override val label: String
        get() = name
}

class BakedKeyMapper(
    val mapper: ExKeyMapper
) : KeyMapper {

    private val strokes: Map<Pitch, KeyStroke> = bake()

    override fun keyof(pitch: Pitch): KeyStroke? = strokes[pitch]

    private fun bake(): Map<Pitch, KeyStroke> =
        mapper.map.entries.associate { entry ->
            pitchOf(entry.key) to KeyStrokeParser.parse(entry.value)
        }

    private fun pitchOf(key: String): Pitch {
        val value: Int = key.trim().toIntOrNull()
            ?: throw IllegalArgumentException("Invalid pitch '$key' in keymap '${mapper.name}'")

        require(value in MIN_PITCH..MAX_PITCH) {
            "Pitch $value is out of range in keymap '${mapper.name}'"
        }

        return Pitch(value)
    }

    private companion object {
        const val MIN_PITCH: Int = 0
        const val MAX_PITCH: Int = 127
    }

}