package com.witelokk.tables

import org.jetbrains.exposed.sql.Table

object Songs: Table("songs") {
    val id = uuid("id").uniqueIndex()
    val name = varchar("name", 255)
    val coverUrl = text("cover_url").nullable()
    val duration = integer("duration")
    val streamUrl = text("stream_url")
    val streamsCount = integer("streams_count")
}
