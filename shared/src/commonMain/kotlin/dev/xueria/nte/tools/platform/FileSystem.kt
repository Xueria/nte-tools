package dev.xueria.nte.tools.platform

interface File {

    fun filename(): String

    fun readText(): String

}

expect fun files(path: String): List<File>
