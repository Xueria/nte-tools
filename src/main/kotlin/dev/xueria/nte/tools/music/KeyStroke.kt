package dev.xueria.nte.tools.music

import java.awt.event.KeyEvent

/**
 * A key combination: one main key plus the modifiers held while it is pressed.
 *
 * Keymap bindings are written as `Z`, `Ctrl+C` or `Ctrl+Shift+F1`.
 */
data class KeyStroke(
    val keyCode: Int,
    val modifiers: Set<KeyModifier> = emptySet()
)

/** A modifier key held for a [KeyStroke]. */
enum class KeyModifier(val keyCode: Int) {
    CONTROL(KeyEvent.VK_CONTROL),
    SHIFT(KeyEvent.VK_SHIFT),
    ALT(KeyEvent.VK_ALT),
    WINDOWS(KeyEvent.VK_WINDOWS)
}

/** Parses key bindings such as `Z`, `Ctrl+C` or `Ctrl+Shift+F1`. */
internal object KeyStrokeParser {

    private const val SEPARATOR: Char = '+'
    private const val SINGLE_CHARACTER: Int = 1

    private val MODIFIERS: Map<String, KeyModifier> = mapOf(
        "ctrl" to KeyModifier.CONTROL,
        "control" to KeyModifier.CONTROL,
        "shift" to KeyModifier.SHIFT,
        "alt" to KeyModifier.ALT,
        "win" to KeyModifier.WINDOWS,
        "windows" to KeyModifier.WINDOWS,
        "super" to KeyModifier.WINDOWS,
        "meta" to KeyModifier.WINDOWS
    )

    private val NAMED_KEYS: Map<String, Int> = buildMap {
        put("space", KeyEvent.VK_SPACE)
        put("enter", KeyEvent.VK_ENTER)
        put("return", KeyEvent.VK_ENTER)
        put("tab", KeyEvent.VK_TAB)
        put("escape", KeyEvent.VK_ESCAPE)
        put("esc", KeyEvent.VK_ESCAPE)
        put("backspace", KeyEvent.VK_BACK_SPACE)
        put("delete", KeyEvent.VK_DELETE)
        put("insert", KeyEvent.VK_INSERT)
        put("home", KeyEvent.VK_HOME)
        put("end", KeyEvent.VK_END)
        put("pageup", KeyEvent.VK_PAGE_UP)
        put("page_up", KeyEvent.VK_PAGE_UP)
        put("pagedown", KeyEvent.VK_PAGE_DOWN)
        put("page_down", KeyEvent.VK_PAGE_DOWN)
        put("up", KeyEvent.VK_UP)
        put("down", KeyEvent.VK_DOWN)
        put("left", KeyEvent.VK_LEFT)
        put("right", KeyEvent.VK_RIGHT)

        (KeyEvent.VK_F1..KeyEvent.VK_F12).forEachIndexed { index, keyCode ->
            put("f${index + 1}", keyCode)
        }
    }

    fun parse(binding: String): KeyStroke {
        val tokens: List<String> = binding.split(SEPARATOR)
            .map { token -> token.trim() }
            .filter { token -> token.isNotEmpty() }

        require(tokens.isNotEmpty()) { "Empty key binding '$binding'" }

        return KeyStroke(
            keyCode = keyCodeOf(tokens.last(), binding),
            modifiers = tokens.dropLast(1).map { token -> modifierOf(token, binding) }.toSet()
        )
    }

    private fun keyCodeOf(token: String, binding: String): Int {
        if (token.length == SINGLE_CHARACTER) {
            return KeyEvent.getExtendedKeyCodeForChar(token[0].uppercaseChar().code)
        }

        return NAMED_KEYS[token.lowercase()]
            ?: throw IllegalArgumentException("Unknown key '$token' in binding '$binding'")
    }

    private fun modifierOf(token: String, binding: String): KeyModifier =
        MODIFIERS[token.lowercase()]
            ?: throw IllegalArgumentException("Unknown modifier '$token' in binding '$binding'")

}