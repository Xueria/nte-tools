package dev.xueria.nte.tools.language

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import dev.xueria.nte.tools.file.FileSystem
import dev.xueria.nte.tools.json.GlobalJson
import dev.xueria.nte.tools.json.JsonFileExtension
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.withContext

object LanguageManager {

    private const val EXTERNAL_LANGUAGE_FOLDER = "language"
    private const val DEFAULT_LOCALE = "zh_cn"

    private val internals = arrayOf(
        "zh_cn", "en_us"
    )

    var languages: Map<String, Language> by mutableStateOf(emptyMap())
        private set

    var language: Language by mutableStateOf(Language.None)
        private set

    fun changeLanguage(locale: String) {
        language = languages[locale] ?: language
    }

    suspend fun loadAll() {
        val catalog = (loadInternalLanguages() + loadExternalLanguages())
            .associateBy { language -> language.locale() }
        languages = catalog
        language = catalog[DEFAULT_LOCALE] ?: catalog.values.firstOrNull() ?: Language.None
    }

    private suspend fun loadInternalLanguages(): List<Language> =
        internals.mapNotNull { filename -> loadInternal(filename) }

    private suspend fun loadExternalLanguages(): List<Language> {
        return withContext(Dispatchers.Default) {
            FileSystem.External.files(EXTERNAL_LANGUAGE_FOLDER)
                .filter { file -> file.extension().equals(JsonFileExtension, true) }
                .map { file ->
                    val src = file.readBytes().first().decodeToString()
                    decodeLanguage(src)
                }
        }
    }

    private suspend fun loadInternal(filename: String): Language? {
        val path = "files/lang/$filename.json"
        return try {
            val file = FileSystem.Internal.file(path)
            val src = file.readBytes().first().decodeToString()
            decodeLanguage(src)
        } catch (exception: CancellationException) {
            throw exception
        } catch (_: Exception) {
            null
        }
    }

    private fun decodeLanguage(src: String): Language =
        GlobalJson.decodeFromString<Language.Default>(src)

}