"""ldap: search and edit an LDAP directory such as Active Directory or OpenLDAP.

Type description, not runnable Starlark: do not `import` anything. The name `ldap` is predeclared in main.star; call `ldap.connect()` directly.
Needs: ldap (connect, search, check_bind) and ldap.write (add, modify, delete), both scoped by server host[:port], for example `ldap:ldap.corp.example.com`. A password given as {"secret": name} also needs secrets. escape_filter and escape_dn need nothing.
Keywords: ldap, active directory, directory, openldap, user lookup, search users, group members, bind, distinguished name, address book

Reading and writing are separate permissions so a lookup extension never gets
write access: `ldap` covers connect, search and check_bind; `ldap.write` covers
add, modify and delete (it also lets the extension connect, but not search).
Use ldaps:// (port 636) or start_tls=True on ldap:// (port 389). The server
certificate is verified against the system roots and there is no option to skip
that. A password is refused over an unencrypted connection unless the host is
this machine.

Search results are a list of dicts: {"dn": "...", "cn": ["Alice"], "mail": [...]}.
Every attribute is a list of values, as LDAP allows several. A value that is
not valid UTF-8 (objectGUID, jpegPhoto) comes back as bytes. Errors from the
server raise; catch with `attempt(conn.search, base, filter)`.

Always pass user input through ldap.escape_filter inside a filter and
ldap.escape_dn inside a DN, never string-concatenate it.

Limits: size_limit defaults to 1000 and is capped at 10000; each operation
times out after the connect timeout (default 15 s); at most 8 connections are
open per extension and they are closed when it stops.
"""

from typing import Any, Dict, List, Optional, Union

Password = Union[str, Dict[str, str], None]
Values = Union[str, bytes, List[Union[str, bytes]]]

class Connection:
    def search(self, base: str, filter: str = "(objectClass=*)", scope: str = "sub", attributes: Optional[List[str]] = None, size_limit: int = 1000) -> List[Dict[str, Any]]:
        """Search below base. scope is "base", "one" or "sub". attributes=None returns all user attributes. Stops at size_limit entries.

        Needs: ldap
        Example: users = conn.search("ou=people,dc=example,dc=com", "(&(objectClass=person)(cn=" + ldap.escape_filter("Ali") + "*))", attributes=["cn", "mail"])
        Task: look up users or groups in a directory
        Keywords: ldap search, find user, lookup, query directory, group members, active directory, filter, memberOf
        """
    def check_bind(self, dn: str, password: Password) -> bool:
        """Check a user's password by binding as them on a separate connection; True when accepted, False for invalid credentials. This connection keeps its own identity. An empty password raises (servers treat it as anonymous).

        Needs: ldap
        Example: ok = conn.check_bind("uid=alice,ou=people,dc=example,dc=com", "her password")
        Task: verify a user's password against the directory
        Keywords: ldap bind, authenticate, login check, verify password, credentials
        """
    def add(self, dn: str, attributes: Dict[str, Values]) -> None:
        """Create an entry. Each attribute value is a string or a list of strings.

        Needs: ldap.write
        Example: conn.add("uid=new,ou=people,dc=example,dc=com", {"objectClass": ["inetOrgPerson"], "uid": "new", "cn": "New", "sn": "User"})
        Task: create a directory entry
        Keywords: ldap add, create user, new entry, provision
        """
    def modify(self, dn: str, replace: Optional[Dict[str, Values]] = None, add: Optional[Dict[str, Values]] = None, delete: Union[Dict[str, Values], List[str], None] = None) -> None:
        """Change an entry. replace sets attributes, add appends values, delete removes the given values (a dict) or whole attributes (a list of names). At least one is required.

        Needs: ldap.write
        Example: conn.modify("uid=alice,ou=people,dc=example,dc=com", replace={"mail": "alice@example.com"}, delete=["description"])
        Task: change attributes of a directory entry
        Keywords: ldap modify, update user, edit attribute, change mail, set attribute
        """
    def delete(self, dn: str) -> None:
        """Delete one entry (it must have no children).

        Needs: ldap.write
        Example: conn.delete("uid=old,ou=people,dc=example,dc=com")
        Task: remove a directory entry
        Keywords: ldap delete, remove user, deprovision
        """
    def close(self) -> None:
        """Close the connection early. Connections are also closed when the extension stops.

        Needs: none
        Example: conn.close()
        """

def connect(url: str, bind_dn: str = "", password: Password = None, start_tls: bool = False, timeout: int = 15) -> Connection:
    """Open a connection to ldap://host[:389] or ldaps://host[:636] and bind as bind_dn (give bind_dn and password together, or neither for an anonymous connection).

    `password` is a string or {"secret": name} to read it from the `secrets` module.
    Needs: ldap scoped to the host (ldap.write also allows connecting)
    Example: conn = ldap.connect("ldaps://ldap.corp.example.com", bind_dn="cn=svc,dc=example,dc=com", password={"secret": "ldap_password"})
    Task: connect to an LDAP or Active Directory server
    Keywords: ldap connect, ldaps, starttls, bind, active directory, directory login, open connection
    """

def escape_filter(value: str) -> str:
    """Escape text for use as a value inside an LDAP search filter, so `*`, `(`, `)`, backslash and NUL cannot change the filter.

    Needs: none
    Example: f = "(uid=" + ldap.escape_filter("a*(b)") + ")"
    Task: put user input safely into an LDAP filter
    Keywords: ldap injection, escape, sanitize filter, quote
    """

def escape_dn(value: str) -> str:
    """Escape text for use as a value inside a distinguished name (comma, plus, quote, backslash, angle brackets, semicolon, and leading or trailing space or #).

    Needs: none
    Example: dn = "cn=" + ldap.escape_dn("Smith, John") + ",ou=people,dc=example,dc=com"
    Task: put user input safely into an LDAP DN
    Keywords: ldap dn, escape, sanitize, distinguished name, quote
    """
