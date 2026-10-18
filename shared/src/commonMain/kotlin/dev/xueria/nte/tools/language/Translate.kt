package dev.xueria.nte.tools.language

import androidx.compose.runtime.mutableStateMapOf
import dev.xueria.nte.tools.Res
import dev.xueria.nte.tools.json.GlobalJson
import dev.xueria.nte.tools.json.JsonFileExtension
import dev.xueria.nte.tools.platform.files

fun translate(translateKey: TranslateKey, defaultText: String? = null): String {
    return defaultText ?: translateKey.toReadableName()
}

object LanguageManager {

    const val LanguageExternalFold = "language"

    val languages = mutableStateMapOf<String, Language>()

    private val internals = arrayOf(
        "en_us", "zh_cn"
    )

    suspend fun loadAll() {
        loadInternal()
        loadExternal()
    }

    private suspend fun loadInternal() {
        for (filename in internals) {
            val bytes = Res.readBytes("files/lang/$filename.json")
            val text = bytes.decodeToString()
            val language = GlobalJson.decodeFromString<Language>(text)
            languages[language.locale] = language
        }
    }

    private fun loadExternal() {
        val files = files(LanguageExternalFold)

        if (files.isEmpty()) {
            return
        }

        files.filter {
            it.extension().equals(JsonFileExtension, ignoreCase = true)
        }.forEach { file ->
            val text = file.readText()
            val language = GlobalJson.decodeFromString<Language>(text)
            languages[language.locale] = language
        }
    }

}