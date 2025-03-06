package com.witelokk.music.models

import kotlinx.serialization.Serializable

@Serializable
data class SearchResult(
    val query: String,
    val page: Int,
    val limit: Int,
    val total: Int,
    val results: List<SearchResultItem>
)
