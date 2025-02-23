package com.witelokk.music.tables

import org.jetbrains.exposed.sql.Table
import org.jetbrains.exposed.sql.jodatime.datetime

object PlaylistSongs : Table("playlist_songs") {
    val playlistId = uuid("playlist_id").references(Playlists.id)
    val songId = uuid("song_id").references(Songs.id)
    val addedAt = datetime("added_at")

    override val primaryKey = PrimaryKey(playlistId, songId)
}