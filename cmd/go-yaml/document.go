// Copyright 2025 The go-yaml Project Contributors
// SPDX-License-Identifier: Apache-2.0

package main

import "go.yaml.in/yaml/v4"

type documentSelection uint8

const (
	documentsAll documentSelection = iota
	documentFirst
	documentLast
)

func selectDocuments[T any](documents []T, selection documentSelection) []T {
	if len(documents) == 0 || selection == documentsAll {
		return documents
	}
	if selection == documentFirst {
		return documents[:1]
	}
	return documents[len(documents)-1:]
}

func selectContractStream[T any](items []T, selection documentSelection,
	typeOf func(T) string,
) []T {
	if selection == documentsAll {
		return items
	}
	var start, end []T
	var documents [][]T
	var document []T
	for _, item := range items {
		switch typeOf(item) {
		case "STREAM-START":
			start = append(start, item)
		case "STREAM-END":
			end = append(end, item)
		case "DOCUMENT-START":
			document = []T{item}
		case "DOCUMENT-END":
			document = append(document, item)
			documents = append(documents, document)
			document = nil
		default:
			if document != nil {
				document = append(document, item)
			}
		}
	}
	selected := selectDocuments(documents, selection)
	result := append([]T{}, start...)
	for _, items := range selected {
		result = append(result, items...)
	}
	return append(result, end...)
}

func selectTokenInfos(infos []*TokenInfo,
	selection documentSelection,
) []*TokenInfo {
	if selection == documentsAll {
		return infos
	}
	var starts []int
	streamEnd := len(infos)
	for i, info := range infos {
		switch info.Token {
		case "DOCUMENT-START":
			starts = append(starts, i)
		case "STREAM-END":
			streamEnd = i
		}
	}
	if len(starts) == 0 {
		return infos
	}
	begins := append([]int(nil), starts...)
	for i := range begins {
		limit := 1
		if i > 0 {
			limit = starts[i-1] + 1
		}
		for begins[i] > limit && tokenPreamble(infos[begins[i]-1]) {
			begins[i]--
		}
	}
	var documents [][]*TokenInfo
	for i, begin := range begins {
		end := streamEnd
		if i+1 < len(begins) {
			end = begins[i+1]
		}
		documents = append(documents, infos[begin:end])
	}
	selected := selectDocuments(documents, selection)
	result := append([]*TokenInfo{}, infos[:1]...)
	for _, document := range selected {
		result = append(result, document...)
	}
	return append(result, infos[streamEnd:]...)
}

func tokenPreamble(info *TokenInfo) bool {
	return info.Token == "VERSION-DIRECTIVE" ||
		info.Token == "TAG-DIRECTIVE" || info.Token == "COMMENT"
}

func selectEventInfos(infos []*EventInfo,
	selection documentSelection,
) []*EventInfo {
	return selectContractStream(infos, selection,
		func(info *EventInfo) string { return info.Event })
}

func selectEvents(events []*Event, selection documentSelection) []*Event {
	return selectContractStream(events, selection,
		func(event *Event) string { return string(event.Type) })
}

func selectNodes(nodes []*yaml.Node, selection documentSelection) []*yaml.Node {
	if selection == documentsAll {
		return nodes
	}
	if len(nodes) == 1 && nodes[0].Kind == yaml.StreamNode {
		stream := *nodes[0]
		stream.Content = selectDocuments(stream.Content, selection)
		return []*yaml.Node{&stream}
	}
	var streamNodes []*yaml.Node
	var documents []*yaml.Node
	for _, node := range nodes {
		if node.Kind == yaml.StreamNode {
			streamNodes = append(streamNodes, node)
		} else {
			documents = append(documents, node)
		}
	}
	if len(streamNodes) == 0 {
		return selectDocuments(documents, selection)
	}
	return append(streamNodes, selectDocuments(documents, selection)...)
}
