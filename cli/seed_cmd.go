package cli

import (
	"context"
	"encoding/json"
	"os"

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
		ctx := context.Background()

		s, err := store.New(ctx, config.Store.Conn)
		if err != nil {
			return err
		}

		bookID := ulid.Make().String()
		person1ID := ulid.Make().String()
		person2ID := ulid.Make().String()
		chapterID := ulid.Make().String()

		err = s.Mutate(ctx,
			store.Mut{
				RecordID: bookID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Publication.Book", map[string]any{
						"title": map[string]string{"eng": "A treatise on nonsense"},
					}),
				},
			},
			store.Mut{
				RecordID: person1ID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Person", nil),
					store.SetAttr("name", "Mr. Whimsi"),
				},
			},
			store.Mut{
				RecordID: person2ID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Person", nil),
					store.SetAttr("name", "Mr. Floppy"),
				},
			},
			store.Mut{
				RecordID: chapterID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Publication.Chapter", nil),
					store.SetAttr("title", "Nonsensical introduction"),
					store.AddRel(ulid.Make().String(), "PartOf", bookID, nil),
					store.AddRel(ulid.Make().String(), "Contribution.Author", person1ID, nil),
					store.AddRel(ulid.Make().String(), "Contribution.Author", person2ID, nil),
				},
			},
			store.Mut{
				RecordID: chapterID,
				Author:   "system",
				Ops: []store.Op{
					store.AddRel(ulid.Make().String(), "Contribution.FirstAuthor", person2ID, nil),
				},
			},
		)
		if err != nil {
			return err
		}

		rec, err := s.GetRec(ctx, chapterID)
		if err != nil {
			return err
		}

		j, _ := json.MarshalIndent(rec, "", "  ")

		os.Stdout.Write(j)

		return nil
	},
}
