package main

import (
	"crypto/tls"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"syscall"

	"github.com/go-ldap/ldap/v3"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
)

var attributes = []string{"cn", "member", "objectClass"}

// Version of the application, e.g. "1.0.0"
var Version = "1.0.0"

// GitCommit is a GIT commit hash
var GitCommit string

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func appendIfMissing(s []*ldap.Entry, i *ldap.Entry) []*ldap.Entry {
	for _, e := range s {
		if e.DN == i.DN {
			return s
		}
	}
	return append(s, i)
}

func entryFinder(l *ldap.Conn, baseDN string, filter string, attributes []string) ([]*ldap.Entry, error) {
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		attributes,
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	return sr.Entries, nil
}

func groupMembers(l *ldap.Conn, groups []*ldap.Entry) ([]*ldap.Entry, error) {
	members := []*ldap.Entry{}

	for _, group := range groups {
		for _, member := range group.GetAttributeValues("member") {
			entries, err := entryFinder(
				l,
				member,
				"(&(distinguishedName="+ldap.EscapeFilter(member)+")(|(objectClass=person)(objectClass=group)))",
				attributes,
			)
			if err != nil {
				return members, err
			}

			// no group members
			if len(entries) == 0 {
				continue
			}

			objectClasses := entries[0].GetAttributeValues("objectClass")
			if contains(objectClasses, "group") {
				entries, err = groupMembers(l, entries)
				if err != nil {
					return members, err
				}
			}

			members = append(members, entries...)
		}
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].GetAttributeValue("cn") < members[j].GetAttributeValue("cn")
	})

	uniqueMembers := []*ldap.Entry{}

	for _, member := range members {
		uniqueMembers = appendIfMissing(uniqueMembers, member)
	}

	return uniqueMembers, nil
}

func main() {

	pflag.CommandLine.SortFlags = false
	host := pflag.StringP("host", "H", "localhost", "LDAP server to query against")
	username := pflag.StringP("username", "u", "", "The full username with domain to bind with (e.g. 'user@example.com')")
	password := pflag.StringP("password", "p", "", "Password to use. If not specified, will be prompted for")
	secure := pflag.Bool("secure", false, "Use LDAPS. This will not verify TLS certs, however. (default: false)")
	baseDN := pflag.StringP("basedn", "b", "", "DN of organizational unit or group to dump members from")
	output := pflag.StringP("output", "o", "", "Save results to file")
	attrs := pflag.StringSlice("attrs", []string{"cn", "mail"}, "Comma separated attributes to dump")
	showVersion := pflag.BoolP("version", "v", false, "Show version and exit")
	pflag.ErrHelp = errors.New("")
	pflag.Parse()

	if *showVersion {
		fmt.Printf("Version: %s\nGitCommit: %s\n", Version, GitCommit)
		os.Exit(0)
	}

	if *username != "" && *password == "" {
		fmt.Fprintf(os.Stderr, "Enter password: ")
		securebytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprint(os.Stderr, "\n")
		*password = string(securebytes)
	}

	attributes = append(attributes, *attrs...)

	proto := "ldap"
	if *secure {
		proto = "ldaps"
	}
	l, err := ldap.DialURL(
		proto+"://"+*host,
		ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
	)
	if err != nil {
		log.Fatal(err)
	}

	if *username == "" {
		err = l.UnauthenticatedBind("")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = l.Bind(*username, *password)
		if err != nil {
			log.Fatal(err)
		}
	}

	defer l.Close()

	groups, err := entryFinder(l, *baseDN, "(&(objectClass=group))", []string{"member"})
	if err != nil {
		log.Fatal(err)
	}

	members, err := groupMembers(l, groups)
	if err != nil {
		log.Fatal(err)
	}

	f := os.Stdout
	if *output != "" {
		f, err = os.Create(*output)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
	}

	w := csv.NewWriter(f)

	for _, member := range members {
		record := []string{}
		for _, attr := range *attrs {
			record = append(record, member.GetAttributeValue(attr))
		}
		if err := w.Write(record); err != nil {
			log.Fatal("error writing record to csv:", err)
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}
