package dev.xueria.nte.tools.file

import kotlinx.coroutines.flow.Flow

interface File {

    fun filename(): String

    fun extension(): String

    fun readBytes(): Flow<ByteArray>

}