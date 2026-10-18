package num

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

const (
	maxPrimeTestBits = 8192
	maxNextPrimeBits = 1024
	maxFactorBits    = 512
	rhoBudget        = 1_200_000
	maxDivisors      = 100000
)

var smallPrimes = func() []int64 {
	const lim = 100000
	sieve := make([]bool, lim+1)
	var ps []int64
	for i := 2; i <= lim; i++ {
		if !sieve[i] {
			ps = append(ps, int64(i))
			for j := i * i; j <= lim; j += i {
				sieve[j] = true
			}
		}
	}
	return ps
}()

func isProbablePrime(n *big.Int) bool { return n.Sign() > 0 && n.ProbablyPrime(20) }

func rho(n *big.Int, budget *int) *big.Int {
	one := big.NewInt(1)
	for c := int64(1); c < 30; c++ {
		cc := big.NewInt(c)
		x, y, q := big.NewInt(2), big.NewInt(2), big.NewInt(1)
		step := func(v *big.Int) { v.Mul(v, v); v.Add(v, cc); v.Mod(v, n) }
		g := big.NewInt(1)
		diff := new(big.Int)
		for g.Cmp(one) == 0 {
			for i := 0; i < 128; i++ {
				step(x)
				step(y)
				step(y)
				diff.Sub(x, y)
				diff.Abs(diff)
				q.Mul(q, diff)
				q.Mod(q, n)
			}
			*budget -= 128
			if *budget <= 0 {
				return nil
			}
			g.GCD(nil, nil, q, n)
		}
		if g.Cmp(n) != 0 {
			return g
		}
	}
	return nil
}

func splitComposite(n *big.Int, budget *int, out *[]*big.Int) error {
	if n.Cmp(big.NewInt(1)) == 0 {
		return nil
	}
	if isProbablePrime(n) {
		*out = append(*out, new(big.Int).Set(n))
		return nil
	}
	d := rho(n, budget)
	if d == nil {
		return errors.New("could not factor completely within the time limit (a large prime factor is left)")
	}
	if err := splitComposite(d, budget, out); err != nil {
		return err
	}
	return splitComposite(new(big.Int).Quo(n, d), budget, out)
}

// factorize returns the prime factors of n >= 2 with multiplicity, ascending.
func factorize(n *big.Int) ([]*big.Int, error) {
	if n.Sign() <= 0 {
		return nil, errors.New("number must be positive")
	}
	if n.BitLen() > maxFactorBits {
		return nil, fmt.Errorf("number is too large to factor (limit %d bits)", maxFactorBits)
	}
	var out []*big.Int
	m := new(big.Int).Set(n)
	r := new(big.Int)
	for _, p := range smallPrimes {
		bp := big.NewInt(p)
		if m.IsInt64() && m.Int64() < p*p {
			break
		}
		for {
			q, rem := new(big.Int).QuoRem(m, bp, r)
			if rem.Sign() != 0 {
				break
			}
			out = append(out, bp)
			m = q
		}
	}
	budget := rhoBudget
	if err := splitComposite(m, &budget, &out); err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Cmp(out[j]) < 0 })
	return out, nil
}

type pe struct {
	p *big.Int
	e int
}

func groupFactors(fs []*big.Int) []pe {
	var out []pe
	for _, f := range fs {
		if len(out) > 0 && out[len(out)-1].p.Cmp(f) == 0 {
			out[len(out)-1].e++
		} else {
			out = append(out, pe{f, 1})
		}
	}
	return out
}

func factorString(n *big.Int) (string, error) {
	if n.Cmp(big.NewInt(1)) == 0 {
		return "1", nil
	}
	fs, err := factorize(n)
	if err != nil {
		return "", err
	}
	var parts []string
	for _, g := range groupFactors(fs) {
		if g.e == 1 {
			parts = append(parts, g.p.String())
		} else {
			parts = append(parts, fmt.Sprintf("%s^%d", g.p, g.e))
		}
	}
	return strings.Join(parts, " * "), nil
}

func divisorsOf(n *big.Int) ([]*big.Int, error) {
	if n.Cmp(big.NewInt(1)) == 0 {
		return []*big.Int{big.NewInt(1)}, nil
	}
	fs, err := factorize(n)
	if err != nil {
		return nil, err
	}
	count := 1
	gs := groupFactors(fs)
	for _, g := range gs {
		count *= g.e + 1
		if count > maxDivisors {
			return nil, fmt.Errorf("more than %d divisors", maxDivisors)
		}
	}
	divs := []*big.Int{big.NewInt(1)}
	for _, g := range gs {
		cur := len(divs)
		pw := big.NewInt(1)
		for k := 1; k <= g.e; k++ {
			pw = new(big.Int).Mul(pw, g.p)
			for i := 0; i < cur; i++ {
				divs = append(divs, new(big.Int).Mul(divs[i], pw))
			}
		}
	}
	sort.Slice(divs, func(i, j int) bool { return divs[i].Cmp(divs[j]) < 0 })
	return divs, nil
}

// nthRoot returns floor(v^(1/k)) for v >= 0.
func nthRoot(v *big.Int, k int) *big.Int {
	if v.Sign() == 0 || k == 1 {
		return new(big.Int).Set(v)
	}
	if k == 2 {
		return new(big.Int).Sqrt(v)
	}
	kk := big.NewInt(int64(k))
	km1 := big.NewInt(int64(k - 1))
	x := new(big.Int).Lsh(big.NewInt(1), uint(v.BitLen()/k+1))
	for {
		pw := new(big.Int).Exp(x, km1, nil)
		y := new(big.Int).Mul(km1, x)
		y.Add(y, new(big.Int).Quo(v, pw))
		y.Quo(y, kk)
		if y.Cmp(x) >= 0 {
			return x
		}
		x = y
	}
}

func rootOf(v *big.Int, k int) (*big.Int, error) {
	if v.Sign() < 0 {
		if k%2 == 0 {
			return nil, errors.New("even root of a negative number")
		}
		r := nthRoot(new(big.Int).Neg(v), k)
		return r.Neg(r), nil
	}
	return nthRoot(v, k), nil
}

func fibPair(n int) (*big.Int, *big.Int) {
	if n == 0 {
		return big.NewInt(0), big.NewInt(1)
	}
	a, b := fibPair(n / 2)
	c := new(big.Int).Mul(a, new(big.Int).Sub(new(big.Int).Lsh(b, 1), a))
	d := new(big.Int).Add(new(big.Int).Mul(a, a), new(big.Int).Mul(b, b))
	if n%2 == 0 {
		return c, d
	}
	return d, new(big.Int).Add(c, d)
}

func isSquare(n *big.Int) bool {
	if n.Sign() < 0 {
		return false
	}
	r := new(big.Int).Sqrt(n)
	return r.Mul(r, r).Cmp(n) == 0
}

func init() {
	one := func(name, title, doc, ret, example string, order int, tags string, f func(n *big.Int) (any, error)) {
		reg(name, title, doc, ret, example, order, tags,
			[]P{rs("value", "Integer, decimal string.")},
			func(a tools.Args) (any, error) {
				n, err := argBig(a, "value", 10)
				if err != nil {
					return nil, err
				}
				return f(n)
			}, accepts(parsesDec), kw(strings.ReplaceAll(strings.TrimPrefix(name, "is_"), "_", " ")))
	}

	one("is_prime", "Is prime", "Tests whether an integer is prime. Exact for numbers below 2^64 and probabilistic (Miller-Rabin plus Baillie-PSW, 20 rounds) above; limited to 8192 bits. Numbers below 2 are not prime.",
		"bool", `num.is_prime("97")`, 30, "prime primality test is prime composite miller rabin", func(n *big.Int) (any, error) {
			if n.BitLen() > maxPrimeTestBits {
				return nil, fmt.Errorf("number is too large (limit %d bits)", maxPrimeTestBits)
			}
			return isProbablePrime(n), nil
		})

	one("next_prime", "Next prime", "Returns the smallest prime strictly greater than the number (next_prime of 13 is 17; of -5 or 0 is 2). Limited to 1024-bit inputs.",
		"str", `num.next_prime("100")`, 31, "next prime following greater prime", func(n *big.Int) (any, error) {
			if n.BitLen() > maxNextPrimeBits {
				return nil, fmt.Errorf("number is too large (limit %d bits)", maxNextPrimeBits)
			}
			if n.Cmp(big.NewInt(2)) < 0 {
				return "2", nil
			}
			c := new(big.Int).Add(n, big.NewInt(1))
			if c.Bit(0) == 0 && c.Cmp(big.NewInt(2)) != 0 {
				c.Add(c, big.NewInt(1))
			}
			for i := 0; i < 100000; i++ {
				if isProbablePrime(c) {
					return c.String(), nil
				}
				c.Add(c, big.NewInt(2))
			}
			return nil, errors.New("no prime found within the search limit")
		})

	one("prev_prime", "Previous prime", "Returns the largest prime strictly smaller than the number (prev_prime of 17 is 13). Numbers 2 and below have no previous prime and give an error. Limited to 1024-bit inputs.",
		"str", `num.prev_prime("100")`, 32, "previous prime smaller lower prime", func(n *big.Int) (any, error) {
			if n.BitLen() > maxNextPrimeBits {
				return nil, fmt.Errorf("number is too large (limit %d bits)", maxNextPrimeBits)
			}
			if n.Cmp(big.NewInt(2)) <= 0 {
				return nil, errors.New("there is no prime below 2")
			}
			if n.Cmp(big.NewInt(3)) == 0 {
				return "2", nil
			}
			c := new(big.Int).Sub(n, big.NewInt(1))
			if c.Bit(0) == 0 {
				c.Sub(c, big.NewInt(1))
			}
			for i := 0; i < 100000 && c.Cmp(big.NewInt(2)) >= 0; i++ {
				if isProbablePrime(c) {
					return c.String(), nil
				}
				c.Sub(c, big.NewInt(2))
			}
			return nil, errors.New("no prime found within the search limit")
		})

	one("prime_factors", "Prime factors", "Returns the prime factors of an integer >= 2 with multiplicity, ascending (360 gives 2, 2, 2, 3, 3, 5). Uses trial division to 100000 and Pollard's rho; numbers over 512 bits or with a huge prime cofactor are an error.",
		"list[str]", `num.prime_factors("360")`, 33, "prime factors factorize factor decomposition integer factorization", func(n *big.Int) (any, error) {
			if n.Cmp(big.NewInt(2)) < 0 {
				return nil, errors.New("number must be at least 2")
			}
			fs, err := factorize(n)
			if err != nil {
				return nil, err
			}
			out := make([]string, len(fs))
			for i, f := range fs {
				out[i] = f.String()
			}
			return out, nil
		})

	one("factorization", "Factorization", "Writes the prime factorization of an integer >= 1 as a product with exponents: 360 gives \"2^3 * 3^2 * 5\", 1 gives \"1\". Same limits as prime_factors.",
		"str", `num.factorization("360")`, 34, "factorization prime power exponent factor tree decomposition", func(n *big.Int) (any, error) {
			if n.Sign() <= 0 {
				return nil, errors.New("number must be at least 1")
			}
			return factorString(n)
		})

	one("divisors", "Divisors", "Lists all positive divisors of an integer >= 1 in ascending order (12 gives 1, 2, 3, 4, 6, 12). Limited to 512-bit numbers and 100000 divisors.",
		"list[str]", `num.divisors("28")`, 35, "divisors factors all divisors proper divisors list", func(n *big.Int) (any, error) {
			if n.Sign() <= 0 {
				return nil, errors.New("number must be at least 1")
			}
			ds, err := divisorsOf(n)
			if err != nil {
				return nil, err
			}
			out := make([]string, len(ds))
			for i, d := range ds {
				out[i] = d.String()
			}
			return out, nil
		})

	one("is_perfect", "Is perfect number", "Tests whether a number equals the sum of its proper divisors (6, 28, 496, 8128, ...). Numbers below 2 are not perfect; limited to 128 bits.",
		"bool", `num.is_perfect("28")`, 36, "perfect number divisors sum aliquot", func(n *big.Int) (any, error) {
			if n.Cmp(big.NewInt(2)) < 0 {
				return false, nil
			}
			if n.BitLen() > 128 {
				return nil, errors.New("number is too large (limit 128 bits)")
			}
			fs, err := factorize(n)
			if err != nil {
				return nil, err
			}
			sigma := big.NewInt(1)
			for _, g := range groupFactors(fs) {
				t := new(big.Int).Exp(g.p, big.NewInt(int64(g.e+1)), nil)
				t.Sub(t, big.NewInt(1))
				t.Quo(t, new(big.Int).Sub(g.p, big.NewInt(1)))
				sigma.Mul(sigma, t)
			}
			return sigma.Cmp(new(big.Int).Lsh(n, 1)) == 0, nil
		})

	one("is_armstrong", "Is Armstrong number", "Tests whether a number equals the sum of its digits each raised to the number of digits (narcissistic number: 153, 370, 9474). Negative numbers are not.",
		"bool", `num.is_armstrong("153")`, 37, "armstrong narcissistic number digits power", func(n *big.Int) (any, error) {
			if n.Sign() < 0 {
				return false, nil
			}
			s := n.String()
			if len(s) > 500 {
				return nil, errors.New("number is too long")
			}
			k := big.NewInt(int64(len(s)))
			sum := new(big.Int)
			for _, c := range s {
				sum.Add(sum, new(big.Int).Exp(big.NewInt(int64(c-'0')), k, nil))
			}
			return sum.Cmp(n) == 0, nil
		})

	one("is_fibonacci", "Is Fibonacci number", "Tests whether a non-negative integer is a Fibonacci number (0, 1, 1, 2, 3, 5, 8, ...) using the 5n^2 +/- 4 perfect-square test.",
		"bool", `num.is_fibonacci("144")`, 38, "fibonacci number test is fibonacci sequence member", func(n *big.Int) (any, error) {
			if n.Sign() < 0 {
				return false, nil
			}
			if n.BitLen() > 100000 {
				return nil, errors.New("number is too large")
			}
			t := new(big.Int).Mul(n, n)
			t.Mul(t, big.NewInt(5))
			return isSquare(new(big.Int).Add(t, big.NewInt(4))) || isSquare(new(big.Int).Sub(t, big.NewInt(4))), nil
		})

	one("fibonacci", "Fibonacci number", "Returns the n-th Fibonacci number (F0=0, F1=1, F10=55) as a decimal string using fast doubling. n must be 0..100000.",
		"str", `num.fibonacci("10")`, 39, "fibonacci nth number sequence golden", func(n *big.Int) (any, error) {
			k, err := smallRange("value", n, 0, 100000)
			if err != nil {
				return nil, err
			}
			f, _ := fibPair(int(k))
			return f.String(), nil
		})

	one("factorial", "Factorial", "Returns n! as a decimal string (0! = 1, 20! = 2432902008176640000). n must be 0..20000; larger values are an error to keep the result small.",
		"str", `num.factorial("20")`, 40, "factorial n! product permutations", func(n *big.Int) (any, error) {
			k, err := smallRange("value", n, 0, 20000)
			if err != nil {
				return nil, err
			}
			if k < 2 {
				return "1", nil
			}
			return new(big.Int).MulRange(1, k).String(), nil
		})

	one("collatz_steps", "Collatz steps", "Counts the steps of the 3n+1 (Collatz) sequence from n >= 1 until it reaches 1 (27 takes 111 steps). Gives an error if it exceeds one million steps.",
		"int", `num.collatz_steps("27")`, 41, "collatz 3n+1 steps hailstone sequence", func(n *big.Int) (any, error) {
			if n.Sign() <= 0 {
				return nil, errors.New("number must be at least 1")
			}
			if n.BitLen() > 4096 {
				return nil, errors.New("number is too large")
			}
			v := new(big.Int).Set(n)
			three := big.NewInt(3)
			for steps := int64(0); steps <= 1000000; steps++ {
				if v.Cmp(big.NewInt(1)) == 0 {
					return steps, nil
				}
				if v.Bit(0) == 0 {
					v.Rsh(v, 1)
				} else {
					v.Mul(v, three)
					v.Add(v, big.NewInt(1))
				}
			}
			return nil, errors.New("sequence is longer than one million steps")
		})

	one("euler_phi", "Euler totient", "Computes Euler's totient phi(n): how many numbers in 1..n are coprime to n (phi(36) = 12, phi(1) = 1). Uses the factorization, so it shares prime_factors' limits.",
		"str", `num.euler_phi("36")`, 42, "euler phi totient coprime count", func(n *big.Int) (any, error) {
			if n.Sign() <= 0 {
				return nil, errors.New("number must be at least 1")
			}
			if n.Cmp(big.NewInt(1)) == 0 {
				return "1", nil
			}
			fs, err := factorize(n)
			if err != nil {
				return nil, err
			}
			phi := new(big.Int).Set(n)
			for _, g := range groupFactors(fs) {
				phi.Quo(phi, g.p)
				phi.Mul(phi, new(big.Int).Sub(g.p, big.NewInt(1)))
			}
			return phi.String(), nil
		})

	one("isqrt", "Integer square root", "Returns floor(sqrt(n)) exactly for any size (isqrt of 99 is 9). Negative input is an error.",
		"str", `num.isqrt("99")`, 43, "integer square root isqrt floor sqrt", func(n *big.Int) (any, error) {
			if n.Sign() < 0 {
				return nil, errors.New("square root of a negative number")
			}
			return new(big.Int).Sqrt(n).String(), nil
		})

	one("icbrt", "Integer cube root", "Returns the cube root truncated toward zero (icbrt of 30 is 3, of -30 is -3), exact for any size.",
		"str", `num.icbrt("1000")`, 44, "integer cube root icbrt floor cbrt", func(n *big.Int) (any, error) {
			r, err := rootOf(n, 3)
			if err != nil {
				return nil, err
			}
			return r.String(), nil
		})

	reg("nth_root", "Integer nth root", "Returns the integer part of the n-th root of a number (nth_root(1000000, 6) is 10), exact for any size. n must be 1..4096; negative values need an odd n and truncate toward zero.",
		"str", `num.nth_root("1024", 10)`, 45, "nth root integer root radical floor exact",
		[]P{rs("value", "Integer, decimal string."), oi("n", "Root degree 1..4096.", 2)},
		func(a tools.Args) (any, error) {
			v, err := argBig(a, "value", 10)
			if err != nil {
				return nil, err
			}
			k, err := intIn("n", a.Int("n"), 1, 4096)
			if err != nil {
				return nil, err
			}
			r, err := rootOf(v, int(k))
			if err != nil {
				return nil, err
			}
			return r.String(), nil
		}, accepts(parsesDec))

	pair := func(name, title, doc, ex, tags string, order int, f func(a, b *big.Int) *big.Int) {
		reg(name, title, doc, "str", ex, order, tags,
			[]P{rs("a", "First integer."), rs("b", "Second integer.")},
			func(x tools.Args) (any, error) {
				a, err := argBig(x, "a", 0)
				if err != nil {
					return nil, err
				}
				b, err := argBig(x, "b", 0)
				if err != nil {
					return nil, err
				}
				return f(a, b).String(), nil
			})
	}
	pair("gcd", "Greatest common divisor", "Greatest common divisor of two integers of any size (gcd(12, 18) = 6; gcd with 0 returns the other absolute value; signs are ignored).",
		`num.gcd("12", "18")`, "gcd greatest common divisor hcf euclid", 46, func(a, b *big.Int) *big.Int { return gcdLcm(false, a, b) })
	pair("lcm", "Least common multiple", "Least common multiple of two integers of any size (lcm(4, 6) = 12; returns 0 if either is 0).",
		`num.lcm("4", "6")`, "lcm least common multiple", 47, func(a, b *big.Int) *big.Int { return gcdLcm(true, a, b) })

	list := func(name, title string, lcm bool) {
		reg(name, title, "Folds a list of integer strings with "+map[bool]string{true: "lcm", false: "gcd"}[lcm]+". The list must not be empty.",
			"str", `num.`+name+`(["12", "18", "30"])`, 48, "gcd lcm list many numbers",
			[]P{rl("values", "Integers as decimal strings.")},
			func(a tools.Args) (any, error) {
				vs := a.List("values")
				if len(vs) == 0 || len(vs) > 10000 {
					return nil, errors.New("values must contain 1..10000 numbers")
				}
				acc := new(big.Int)
				for i, s := range vs {
					n, err := parseBig(s, 0)
					if err != nil {
						return nil, fmt.Errorf("values[%d]: %w", i, err)
					}
					if i == 0 {
						acc.Abs(n)
					} else {
						acc = gcdLcm(lcm, acc, n)
					}
				}
				return acc.String(), nil
			})
	}
	list("gcd_list", "GCD of a list", false)
	list("lcm_list", "LCM of a list", true)

	reg("modpow", "Modular exponentiation", "Computes base^exp mod m with arbitrary precision (modpow(4, 13, 497) = 445). m must be positive; a negative exp needs base invertible mod m.",
		"str", `num.modpow("4", "13", "497")`, 49, "modpow modular exponentiation pow mod power modulo rsa",
		[]P{rs("base", "Base."), rs("exp", "Exponent."), rs("mod", "Modulus, > 0.")},
		func(a tools.Args) (any, error) {
			b, err := argBig(a, "base", 0)
			if err != nil {
				return nil, err
			}
			e, err := argBig(a, "exp", 0)
			if err != nil {
				return nil, err
			}
			m, err := argBig(a, "mod", 0)
			if err != nil {
				return nil, err
			}
			if m.Sign() <= 0 {
				return nil, errors.New("mod must be positive")
			}
			if m.BitLen() > 1<<16 || e.BitLen() > 1<<16 {
				return nil, errors.New("operands are too large")
			}
			r := new(big.Int).Exp(b, e, m)
			if r == nil {
				return nil, errors.New("base has no inverse for the negative exponent")
			}
			return r.String(), nil
		})

	reg("modinv", "Modular inverse", "Finds x with a*x = 1 (mod m), 0 <= x < m (modinv(3, 11) = 4). Errors when gcd(a, m) is not 1 or m < 2.",
		"str", `num.modinv("3", "11")`, 50, "modular inverse modinv multiplicative inverse extended euclid",
		[]P{rs("a", "Number to invert."), rs("m", "Modulus, >= 2.")},
		func(a tools.Args) (any, error) {
			x, err := argBig(a, "a", 0)
			if err != nil {
				return nil, err
			}
			m, err := argBig(a, "m", 0)
			if err != nil {
				return nil, err
			}
			if m.Cmp(big.NewInt(2)) < 0 {
				return nil, errors.New("m must be at least 2")
			}
			r := new(big.Int).ModInverse(x, m)
			if r == nil {
				return nil, errors.New("no inverse: a and m are not coprime")
			}
			return r.String(), nil
		})
}
