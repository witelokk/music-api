package com.witelokk.tables

import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.jodatime.date

enum class ReleaseTypes {
    single, ep, album,
}

object Releases: Table("releases") {
    val id = uuid("id").uniqueIndex()
    val name = varchar("name", 255)
    val coverUrl = text("cover_url").nullable()
    val type = enumeration("type", ReleaseTypes::class)
    val releasedAt = date("release_at")
}
