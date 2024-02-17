package cli

import (
	"context"

	"github.com/oklog/ulid/v2"
	"github.com/spf13/cobra"
	"github.com/ugent-library/muts/store"
)

func init() {
	rootCmd.AddCommand(seedCmd)
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := store.New(context.Background(), config.Store.Conn)
		if err != nil {
			return err
		}

		bookID := ulid.Make().String()
		author1ID := ulid.Make().String()
		author2ID := ulid.Make().String()
		chapterID := ulid.Make().String()

		return s.Mutate(context.Background(),
			store.Mut{
				RecordID: bookID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Publication.Book"),
					store.AddAttr("title.eng", "A treatise on nonsense"),
				},
			},
			store.Mut{
				RecordID: author1ID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Contributor"),
					store.AddAttr("name", "Mr. Whimsi"),
				},
			},
			store.Mut{
				RecordID: author2ID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Contributor"),
					store.AddAttr("name", "Mr. Floppy"),
				},
			},
			store.Mut{
				RecordID: chapterID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Publication.Chapter"),
					store.AddAttr("title.eng", "Nonsensical introduction"),
					store.AddRel("partOf", bookID),
					store.AddRel("hasAuthor", author1ID),
					store.AddRel("hasFirstAuthor", author1ID),
					store.AddRel("hasAuthor", author2ID),
				},
			},
		)
	},
}
