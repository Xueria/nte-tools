package dev.xueria.nte.tools.file

import dev.xueria.nte.tools.Res
import dev.xueria.nte.tools.platform.RunningPath
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOf
import java.nio.file.Path
import kotlin.io.path.Path
import kotlin.io.path.exists
import kotlin.io.path.extension
import kotlin.io.path.getLastModifiedTime
import kotlin.io.path.isRegularFile
import kotlin.io.path.listDirectoryEntries
import kotlin.io.path.name
import kotlin.io.path.readBytes

sealed interface FileSystem {

    fun file(path: String): File

    object Internal : FileSystem {

        override fun file(path: String): File {
            val fake = Path(path)
            return object : File {
                override fun filename(): String = fake.name

                override fun extension(): String = fake.extension

                override fun lastModified(): Long = 0

                override fun readBytes(): Flow<ByteArray> = flow {
                    emit(Res.readBytes(path))
                }
            }
        }

    }

    object External : FileSystem {

        override fun file(path: String): File {
            val file = Path(RunningPath).resolve(path)
            return file(file)
        }

        private fun file(file: Path): File {
            return object : File {
                override fun filename(): String = file.name

                override fun extension(): String = file.extension

                override fun lastModified(): Long =
                    runCatching { file.getLastModifiedTime().toMillis() }.getOrDefault(0)

                override fun readBytes(): Flow<ByteArray> = flowOf(file.readBytes())
            }
        }

        fun files(directory: String): List<File> {
            val directory = Path(RunningPath).resolve(directory)
            if (!directory.exists()) {
                return emptyList()
            }
            return directory.listDirectoryEntries()
                .filter { it.isRegularFile() }
                .map { file(it) }
        }

    }

}