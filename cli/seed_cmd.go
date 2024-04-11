package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
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

		poolConfig, err := pgxpool.ParseConfig(config.Store.Conn)
		if err != nil {
			return err
		}
		// poolConfig.ConnConfig.Tracer = pgxslog.NewTracer(logger)
		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			return err
		}
		s, err := store.NewFromPool(pool)
		if err != nil {
			return err
		}

		bookID := ulid.Make().String()
		person1ID := ulid.Make().String()
		person2ID := ulid.Make().String()
		chapterID := ulid.Make().String()
		orgIDs := []string{
			ulid.Make().String(),
			ulid.Make().String(),
			ulid.Make().String(),
		}

		err = s.Mutate(ctx,
			store.Mut{
				RecordID: orgIDs[0],
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Organization", map[string]any{"name": "UGent"}),
				},
			},
			store.Mut{
				RecordID: orgIDs[1],
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Organization", map[string]any{"name": "UGent/CA"}),
					store.AddRel(ulid.Make().String(), "Parent", orgIDs[0], nil),
				},
			},
			store.Mut{
				RecordID: orgIDs[2],
				Author:   "system",
				Ops: []store.Op{
					store.AddRec("Organization", map[string]any{"name": "UGent/CA/CA20"}),
					store.AddRel(ulid.Make().String(), "Parent", orgIDs[1], nil),
				},
			},
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
					store.AddRel(ulid.Make().String(), "Affiliation", orgIDs[2], nil),
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
					store.AddRel(ulid.Make().String(), "Contribution.Author", person2ID, map[string]any{
						"creditRole": "firstAuthor",
					}),
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

		recs, err := s.Many(ctx, store.Query{Limit: 2})
		if err != nil {
			return err
		}
		j, _ := json.Marshal(recs)
		fmt.Printf("many limit 2: %s\n", j)

		recs, err = s.Many(ctx, store.Query{Limit: 10, ID: chapterID, Follow: "PartOf|Contribution.*"})
		if err != nil {
			return err
		}
		j, _ = json.Marshal(recs)
		fmt.Printf("id + follow: %s\n", j)

		recs, err = s.Many(ctx, store.Query{Limit: 10, IDIn: []string{chapterID, bookID}})
		if err != nil {
			return err
		}
		j, _ = json.Marshal(recs)
		fmt.Printf("many id in: %s\n", j)

		recs, err = s.Many(ctx, store.Query{Limit: 10, Kind: "Publication.Chapter"})
		if err != nil {
			return err
		}
		j, _ = json.Marshal(recs)
		fmt.Printf("many kind Publication.Chapter: %s\n", j)

		recs, err = s.Many(ctx, store.Query{Limit: 10, Kind: "Publication.*", Follow: "PartOf"})
		if err != nil {
			return err
		}
		j, _ = json.Marshal(recs)
		fmt.Printf("many kind Publication.*: %s\n", j)

		recs, err = s.Many(ctx, store.Query{Limit: 10, Attr: "$.title"})
		if err != nil {
			return err
		}
		j, _ = json.Marshal(recs)
		fmt.Printf("many attr $.title: %s\n", j)

		return nil
	},
}
