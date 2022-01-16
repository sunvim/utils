package main

import (
	"fmt"

	"github.com/sunvim/utils/patricia"
)

func main() {
	// Create a new tree.
	trie := patricia.NewTrie()

	// Insert some items.
	trie.Insert(patricia.Prefix("Pepa Novak"), 1)
	trie.Insert(patricia.Prefix("Pepa Sindelar"), 2)
	trie.Insert(patricia.Prefix("Karel Macha"), 3)
	trie.Insert(patricia.Prefix("Karel Hynek Macha"), 4)

	// Just check if some things are present in the tree.
	key := patricia.Prefix("Pepa Novak")
	fmt.Printf("%q present? %v\n", key, trie.Match(key))
	key = patricia.Prefix("Karel")
	fmt.Printf("Anybody called %q here? %v\n", key, trie.MatchSubtree(key))

	// Walk the tree.
	trie.Visit(printItem)
	// "Karel Hynek Macha": 4
	// "Karel Macha": 3
	// "Pepa Novak": 1
	// "Pepa Sindelar": 2

	// Walk a subtree.
	trie.VisitSubtree(patricia.Prefix("Pepa"), printItem)
	// "Pepa Novak": 1
	// "Pepa Sindelar": 2

	// Modify an item, then fetch it from the tree.
	trie.Set(patricia.Prefix("Karel Hynek Macha"), 10)
	key = patricia.Prefix("Karel Hynek Macha")
	fmt.Printf("%q: %v\n", key, trie.Get(key))
	// "Karel Hynek Macha": 10

	// Walk prefixes.
	prefix := patricia.Prefix("Karel Hynek Macha je kouzelnik")
	trie.VisitPrefixes(prefix, printItem)
	// "Karel Hynek Macha": 10

	// Delete some items.
	trie.Delete(patricia.Prefix("Pepa Novak"))
	trie.Delete(patricia.Prefix("Karel Macha"))

	// Walk again.
	trie.Visit(printItem)
	// "Karel Hynek Macha": 10
	// "Pepa Sindelar": 2

	// Delete a subtree.
	trie.DeleteSubtree(patricia.Prefix("Pepa"))

	// Print what is left.
	trie.Visit(printItem)
	// "Karel Hynek Macha": 10

}

func printItem(prefix patricia.Prefix, item patricia.Item) error {
	fmt.Printf("%s  %v\n", prefix, item)
	return nil
}
