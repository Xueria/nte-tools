package dev.xueria.nte.tools.music

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import dev.xueria.nte.tools.file.File
import dev.xueria.nte.tools.file.FileStamp
import dev.xueria.nte.tools.file.FileSystem
import dev.xueria.nte.tools.file.stamp
import dev.xueria.nte.tools.json.GlobalJson
import dev.xueria.nte.tools.json.JsonFileExtension
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext

/**
 * Loads the keymaps shipped inside the application and the ones dropped next to it.
 *
 * The built-in `k21` layout is read from `files/music/k21.json`. Every `*.json`
 * file in `<working directory>/music/map` is loaded on top of it, so an external
 * file with the same name replaces the built-in layout.
 *
 * [refresh] is cheap enough to be polled: it only re-reads the external files
 * when the folder content changed.
 */
object MusicKeyMaps {

    private const val INTERNAL_FOLDER = "files/music"
    private const val EXTERNAL_FOLDER = "music/map"
    private const val INTERNAL_NAME = "k21"

    private val mutex: Mutex = Mutex()

    /** The packaged layouts, read once because a resource cannot change at runtime. */
    private var internals: List<ExKeyMapper> = emptyList()

    /** Modification stamps of the last scan; `null` forces the next reload. */
    private var signature: List<FileStamp>? = null

    var keyMaps: Map<String, ExKeyMapper> by mutableStateOf(emptyMap())
        private set

    suspend fun loadAll() {
        mutex.withLock {
            internals = loadInternals()
            signature = null
            reload()
        }
    }

    /** Reloads the external keymaps when the map folder changed since the last scan. */
    suspend fun refresh() {
        mutex.withLock { reload() }
    }

    private suspend fun reload() {
        val files: List<File> = scan()
        val current: List<FileStamp> = files.map { file -> file.stamp() }

        if (current == signature) return

        signature = current
        keyMaps = (internals + parse(files)).associateBy { mapper -> mapper.name }
    }

    private suspend fun loadInternals(): List<ExKeyMapper> =
        listOfNotNull(load(FileSystem.Internal.file("$INTERNAL_FOLDER/$INTERNAL_NAME.$JsonFileExtension")))

    private suspend fun scan(): List<File> =
        withContext(Dispatchers.Default) {
            FileSystem.External.files(EXTERNAL_FOLDER)
                .filter { file -> file.extension().equals(JsonFileExtension, true) }
        }

    private suspend fun parse(files: List<File>): List<ExKeyMapper> =
        withContext(Dispatchers.Default) {
            files.mapNotNull { file -> load(file) }
        }

    private suspend fun load(file: File): ExKeyMapper? =
        try {
            decode(file.readBytes().first().decodeToString())
        } catch (exception: CancellationException) {
            throw exception
        } catch (_: Exception) {
            null
        }

    private fun decode(src: String): ExKeyMapper = GlobalJson.decodeFromString<ExKeyMapper>(src)

}
