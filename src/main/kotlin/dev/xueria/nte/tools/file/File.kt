package dev.xueria.nte.tools.file

import kotlinx.coroutines.flow.Flow

interface File {

    fun filename(): String

    fun extension(): String

    /** Last modification time in milliseconds, or zero when the source has no clock. */
    fun lastModified(): Long

    fun readBytes(): Flow<ByteArray>

}

/** Identity of one file on disk, used to detect added, removed or edited files. */
data class FileStamp(val name: String, val modified: Long)

/** Takes a snapshot of this file's identity. */
fun File.stamp(): FileStamp = FileStamp(filename(), lastModified())