package dev.xueria.nte.tools.music

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import dev.xueria.nte.tools.file.File
import dev.xueria.nte.tools.file.FileStamp
import dev.xueria.nte.tools.file.FileSystem
import dev.xueria.nte.tools.file.stamp
import dev.xueria.nte.tools.music.parse.MusicFileDispatcher
import dev.xueria.nte.tools.music.parse.ParseInput
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import kotlinx.coroutines.withContext

/**
 * Loads the sheets of the working directory and keeps them for the player.
 *
 * Sheets live in `<working directory>/music`; the `music/map` subdirectory is
 * skipped on purpose because it holds keymaps, not sheets.
 *
 * [refresh] is cheap enough to be polled: it only re-parses the files when the
 * folder content changed.
 */
object Music {

    private const val EXTERNAL_FOLDER = "music"
    private const val MIDI_TYPE = "midi"

    private val mutex: Mutex = Mutex()

    /** Modification stamps of the last scan; `null` forces the next reload. */
    private var signature: List<FileStamp>? = null

    var sheets: Map<String, MusicSheet> by mutableStateOf(emptyMap())
        private set

    suspend fun loadAll() {
        MusicFileDispatcher.buildIn()
        MusicKeyMaps.loadAll()

        signature = null
        refresh()
    }

    /** Reloads the sheets when the music folder changed since the last scan. */
    suspend fun refresh() {
        mutex.withLock {
            val files: List<File> = scan()
            val current: List<FileStamp> = files.map { file -> file.stamp() }

            if (current == signature) return

            signature = current
            sheets = parse(files)
        }
    }

    private suspend fun scan(): List<File> =
        withContext(Dispatchers.Default) {
            FileSystem.External.files(EXTERNAL_FOLDER)
        }

    private suspend fun parse(files: List<File>): Map<String, MusicSheet> =
        withContext(Dispatchers.Default) {
            files.mapNotNull { file -> load(file) }
                .sortedBy { entry -> entry.id }
                .associate { entry -> entry.id to entry.sheet }
        }

    private suspend fun load(file: File): SheetFile? {
        val type: String = typeOf(file) ?: return null

        return try {
            val result: Result<MusicSheet> = MusicFileDispatcher.dispatch(
                ParseInput(bytes = file.readBytes().first(), type = type)
            )

            result.getOrNull()?.let { sheet -> SheetFile(idOf(file), sheet) }
        } catch (exception: CancellationException) {
            throw exception
        } catch (_: Exception) {
            null
        }
    }

    /** Maps a file extension to the type understood by [MusicFileDispatcher]. */
    private fun typeOf(file: File): String? =
        when (file.extension().lowercase()) {
            "mid", "midi" -> MIDI_TYPE
            else -> null
        }

    private fun idOf(file: File): String = file.filename().substringBeforeLast('.')

}

private data class SheetFile(val id: String, val sheet: MusicSheet)
