package dev.xueria.nte.tools.language

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import dev.xueria.nte.tools.Res
import dev.xueria.nte.tools.json.GlobalJson
import dev.xueria.nte.tools.json.JsonFileExtension
import dev.xueria.nte.tools.platform.files
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

object LanguageManager {

    private const val EXTERNAL_LANGUAGE_FOLDER = "language"
    private const val DEFAULT_LOCALE = "zh_cn"

    private val internals = arrayOf(
        "en_us", "zh_cn"
    )

    var languages by mutableStateOf<Map<String, Language>>(emptyMap())
        private set

    var language by mutableStateOf<Language?>(null)
        private set

    fun changeLanguage(locale: String) {
        language = languages[locale] ?: language
    }

    fun translate(translateKey: TranslateKey): String? {
        return language?.messages?.get(translateKey)
    }

    suspend fun loadAll() {
        val catalog = (loadInternalLanguages() + loadExternalLanguages())
            .associateBy { language -> language.locale }
        languages = catalog
        changeLanguage(DEFAULT_LOCALE)
    }

    private suspend fun loadInternalLanguages(): List<Language> =
        internals.mapNotNull { filename -> loadInternal(filename) }

    private suspend fun loadExternalLanguages(): List<Language> {
        return withContext(Dispatchers.Default) {
            files(EXTERNAL_LANGUAGE_FOLDER)
                .filter { file -> file.extension().equals(JsonFileExtension, true) }
                .map { file ->
                    decodeLanguage(file.readText())
                }
        }
    }

    private suspend fun loadInternal(filename: String): Language? {
        val path = "files/lang/$filename.json"
        return try {
            decodeLanguage(Res.readBytes(path).decodeToString())
        } catch (exception: CancellationException) {
            throw exception
        } catch (_: Exception) {
            null
        }
    }

    private fun decodeLanguage(src: String): Language =
        GlobalJson.decodeFromString<Language>(src)

}