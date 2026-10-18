package dev.xueria.nte.tools.platform

import java.nio.file.Path
import kotlin.io.path.Path
import kotlin.io.path.exists
import kotlin.io.path.extension
import kotlin.io.path.isRegularFile
import kotlin.io.path.listDirectoryEntries
import kotlin.io.path.name
import kotlin.io.path.readText

private class FileImplement(private val file: Path) : File {

    override fun filename(): String = file.name

    override fun extension(): String = file.extension

    override fun readText(): String = file.readText()

}

actual fun files(path: String): List<File> {
    val directory = Path(path)

    if (!directory.exists()) {
        return emptyList()
    }

    return directory.listDirectoryEntries()
        .filter { it.isRegularFile() }
        .map { FileImplement(it) }
}