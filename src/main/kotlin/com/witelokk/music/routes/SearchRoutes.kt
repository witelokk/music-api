package com.witelokk.music.routes

import com.witelokk.music.models.*
import com.witelokk.music.tables.*
import com.witelokk.music.tables.Releases
import com.witelokk.music.tables.Songs
import io.ktor.http.*
import io.ktor.server.response.*
import io.ktor.server.auth.*
import io.ktor.server.auth.jwt.*
import io.ktor.server.routing.*
import io.github.smiley4.ktoropenapi.route
import io.github.smiley4.ktoropenapi.get
import org.jetbrains.exposed.sql.*
import org.jetbrains.exposed.sql.transactions.transaction
import java.util.*


fun Route.searchRoutes() {
    authenticate("auth-jwt") {
        route("/search", {
            tags = listOf("search")
        }) {

            get("", {
                description = "Search for artists, songs, albums, and playlists"
                request {
                    queryParameter<String>("q") {
                        description = "Search query"
                    }
                    queryParameter<String>("type") {
                        description = "Search type"
                    }
                    queryParameter<String>("page") {
                        description = "Page number"
                    }
                    queryParameter<String>("limit") {
                        description = "Results per page"
                    }
                }
                response {
                    HttpStatusCode.OK to {
                        description = "Success"
                        body<SearchResult>()
                    }
                }
            }) {
                val principal = call.principal<JWTPrincipal>()
                val userId = UUID.fromString(principal!!.payload.getClaim("sub").asString())

                val query = call.request.queryParameters["q"] ?: return@get call.respond(
                    HttpStatusCode.BadRequest, FailureResponse("missing_query", "Missing query parameter")
                )
                val type = call.request.queryParameters["type"]
                val page = call.request.queryParameters["page"]?.toIntOrNull() ?: 1
                val limit = call.request.queryParameters["limit"]?.toIntOrNull() ?: 10

                val results = search(userId, query, type, page, limit)

                call.respond(results)
            }
        }
    }
}

private fun search(userId: UUID, query: String, type: String?, page: Int, limit: Int): SearchResult {
    val offset = (page - 1) * limit

    return transaction {
        val favoriteSongIds = Favorites.select { Favorites.userId eq userId }
            .map { it[Favorites.songId] }
            .toSet()

        val songResults = if (type == null || type == "song") {
            Songs.select { Songs.name ilike "%$query%" }
                .map { row ->
                    val songId = row[Songs.id]

                    val artists = SongArtists.innerJoin(Artists)
                        .select { SongArtists.songId eq songId }
                        .map {
                            AristSummary(
                                id = it[Artists.id],
                                name = it[Artists.name],
                                avatarUrl = it[Artists.avatarUrl]
                            )
                        }

                    SearchResultItem(
                        type = "song",
                        song = Song(
                            id = songId,
                            name = row[Songs.name],
                            coverUrl = row[Songs.coverUrl],
                            durationSeconds = row[Songs.duration],
                            streamUrl = row[Songs.streamUrl],
                            isFavorite = songId in favoriteSongIds,
                            artists = artists
                        )
                    )
                }
        } else emptyList()

        val releaseResults = if (type == null || type == "release") {
            Releases.select { Releases.name ilike "%$query%" }
                .mapNotNull { row ->
                    val releaseId = row[Releases.id]

                    val songCount = Songs.innerJoin(ReleaseSongs)
                        .select { ReleaseSongs.releaseId eq releaseId }
                        .count()

                    if (songCount <= 1) return@mapNotNull null

                    val artists = ReleaseArtists.innerJoin(Artists)
                        .select { ReleaseArtists.releaseId eq releaseId }
                        .map {
                            AristSummary(
                                id = it[Artists.id],
                                name = it[Artists.name],
                                avatarUrl = it[Artists.avatarUrl]
                            )
                        }

                    SearchResultItem(
                        type = "release",
                        release = ReleaseSummary(
                            id = releaseId,
                            name = row[Releases.name],
                            coverUrl = row[Releases.coverUrl],
                            type = row[Releases.type].name,
                            releasedAt = row[Releases.releasedAt],
                            artists = ArtistsSummary(artists.size, artists)
                        )
                    )
                }
        } else emptyList()

        val artistResults = if (type == null || type == "artist") {
            Artists.select { Artists.name ilike "%$query%" }
                .map { row ->
                    SearchResultItem(
                        type = "artist",
                        artist = AristSummary(
                            id = row[Artists.id],
                            name = row[Artists.name],
                            avatarUrl = row[Artists.avatarUrl]
                        )
                    )
                }
        } else emptyList()

        val playlistResults = if (type == null || type == "playlist") {
            Playlists.select { (Playlists.name ilike "%$query%") and (Playlists.userId eq userId) }
                .map { row ->
                    val playlistId = row[Playlists.id]

                    val songsCount = Songs.innerJoin(PlaylistSongs)
                        .select { PlaylistSongs.playlistId eq playlistId }
                        .count()

                    SearchResultItem(
                        type = "playlist",
                        playlist = PlaylistSummary(
                            id = playlistId,
                            name = row[Playlists.name],
                            coverUrl = null,
                            songsCount = songsCount.toInt()
                        )
                    )
                }
        } else emptyList()

        val allResults = (songResults + releaseResults + artistResults + playlistResults)
            .sortedBy { it.type }
            .drop(offset)
            .take(limit)

        SearchResult(
            query = query,
            page = page,
            limit = limit,
            total = allResults.size,
            results = allResults
        )
    }
}

class ILikeOp(expr1: Expression<*>, expr2: Expression<*>) : ComparisonOp(expr1, expr2, "ILIKE")

infix fun <T : String?> ExpressionWithColumnType<T>.ilike(pattern: String): Op<Boolean> =
    ILikeOp(this, QueryParameter(pattern, columnType))
