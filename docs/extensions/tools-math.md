# Math and statistics (`math`)

An expression evaluator, percentages, statistics, algebra, geometry and finance helpers.

No permission is needed: these are pure functions. Generated from the tools registry.

116 functions. Use them as `math.<name>(...)` in any extension.

## math.eval_str

`math.eval_str(expr: str, degrees: bool = False) -> str`

Evaluates a math expression (same syntax as eval) and returns a clean string with up to 12 significant digits and no trailing zeros, so 0.1+0.2 gives "0.3" and 1/3 gives "0.333333333333". Very large or small results use exponent form (1e+25).

| Parameter | Type | Description |
|---|---|---|
| `expr` | `str` | The expression, e.g. 2*(3+4)^2, sqrt(16), 200 + 10%, 15% of 80. |
| `degrees` | `bool` (default `False`) | Trig functions take and return degrees instead of radians. |

Returns `str`. Example: `math.eval_str("0.1 + 0.2")`

## math.eval

`math.eval(expr: str, degrees: bool = False) -> float`

Evaluates a math expression with floats and returns the number. Supports + - * / % (remainder) ^ or ** (power), parentheses, implicit multiplication (2(3+4), 3pi), constants pi e tau phi, factorial 5!, thousands separators (1,000), scientific notation (1e3), functions sin cos tan asin acos atan atan2 sinh cosh tanh sqrt cbrt exp ln log(x) \[base 10\] log(x,base) log2 log10 abs floor ceil round(x\[,d\]) trunc sign min max sum avg pow hypot gamma factorial deg rad, and the word 'of' as multiplication. Percent: 50% is 0.5, and in a sum 'a + b%' means a plus b percent of a (200 + 10% = 220); a comma is never a decimal point. Errors name the character position.

| Parameter | Type | Description |
|---|---|---|
| `expr` | `str` | The expression, e.g. 2*(3+4)^2, sqrt(16), 200 + 10%, 15% of 80. |
| `degrees` | `bool` (default `False`) | Trig functions take and return degrees instead of radians. |

Returns `float`. Example: `math.eval("2(3+4) + 10%")`

## math.percent_text

`math.percent_text(text: str) -> Dict[str, str]`

Answers a percentage question typed in words. Understands "15% of 80", "20 is what % of 50", "what percent is 20 of 50", "50 is 25% of what", "from 80 to 100" (percent change) and "200 + 10%" / "200 - 10%". Returns a dict with result plus an explanation.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | The question, e.g. 15% of 80. |

Returns `Dict[str, str]`. Example: `math.percent_text("15% of 80")`

## math.percent_of

`math.percent_of(p: float, x: float) -> float`

Returns p percent of x (percent_of(15, 80) = 12).

| Parameter | Type | Description |
|---|---|---|
| `p` | `float` | p value. |
| `x` | `float` | x value. |

Returns `float`. Example: `math.percent_of(15, 80)`

## math.stats_text

`math.stats_text(text: str) -> Dict[str, str]`

Extracts every number from any text (commas, newlines or words in between; 1,234 is one number) and returns summary statistics like describe. Errors if the text has no numbers.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text containing numbers. |

Returns `Dict[str, str]`. Example: `math.stats_text("scores: 3, 5, 8 and 13")`

## math.describe

`math.describe(values: List[float]) -> Dict[str, str]`

Summary statistics as a dict: count, sum, mean, median, mode, min, max, range, sample variance and stdev, pop_variance, pop_stdev, q1, q3, iqr and (for positive data) geometric_mean and harmonic_mean.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `Dict[str, str]`. Example: `math.describe([2, 4, 4, 4, 5, 5, 7, 9])`

## math.loan_payment

`math.loan_payment(principal: float, annual_rate: float, years: float, per_year: int = 12) -> Dict[str, str]`

Fixed payment of an amortising loan or mortgage: principal, annual rate in percent, term in years and payments per year (default 12). Returns payment, total repaid, total interest and the number of payments, rounded to cents (200000 at 6% for 30 years gives 1199.10 a month).

| Parameter | Type | Description |
|---|---|---|
| `principal` | `float` | Amount borrowed. |
| `annual_rate` | `float` | Annual interest rate in percent (0 allowed). |
| `years` | `float` | Term in years. |
| `per_year` | `int` (default `12`) | Payments per year. |

Returns `Dict[str, str]`. Example: `math.loan_payment(200000, 6, 30)`

## math.tip_text

`math.tip_text(text: str) -> Dict[str, str]`

Reads "bill \[tip%\] \[people\]" from text ("85.50 18% 2" or "85.5 18 2"; tip defaults to 15%, people to 1) and returns the same dict as tip.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Bill, optional tip percent, optional number of people. |

Returns `Dict[str, str]`. Example: `math.tip_text("85.50 18% 2")`

## math.loan_text

`math.loan_text(text: str) -> Dict[str, str]`

Reads "principal rate% years \[payments_per_year\]" from text ("200000 6% 30") and returns the same dict as loan_payment: payment, total, interest and payments.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Principal, annual rate percent, years, optional payments per year. |

Returns `Dict[str, str]`. Example: `math.loan_text("200000 6% 30")`

## math.solve_quadratic

`math.solve_quadratic(a: float, b: float, c: float) -> List[str]`

Solves a*x^2 + b*x + c = 0 and returns the roots as strings, real ones ascending (a double root once) and complex ones like "1+2i", "1-2i". With a=0 it solves the linear equation. x^2-3x+2 gives \["1", "2"\].

| Parameter | Type | Description |
|---|---|---|
| `a` | `float` | Coefficient of x^2. |
| `b` | `float` | Coefficient of x. |
| `c` | `float` | Constant term. |

Returns `List[str]`. Example: `math.solve_quadratic(1, -3, 2)`

## math.quadratic_text

`math.quadratic_text(text: str) -> Dict[str, str]`

Solves a quadratic written as text, either "x^2 - 3x + 2 = 0" or three coefficients "1 -3 2". Returns a, b, c, discriminant and the roots x1, x2 (complex roots as strings).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Equation or coefficients a b c. |

Returns `Dict[str, str]`. Example: `math.quadratic_text("x^2 - 3x + 2 = 0")`

## math.fraction_calc

`math.fraction_calc(expr: str) -> str`

Evaluates an expression exactly with fractions: numbers, decimals, + - * / and parentheses ("3/4 + 1/6" gives "11/12"). Division is exact, never floating point.

| Parameter | Type | Description |
|---|---|---|
| `expr` | `str` | Expression such as 3/4 + 1/6 or (1/2)*(2/3). |

Returns `str`. Example: `math.fraction_calc("3/4 + 1/6")`

## math.time_sum

`math.time_sum(expr: str) -> str`

Adds and subtracts clock times: "1:30 + 2:45" gives "4:15"; also h:mm:ss and unit forms "1h 30m + 45m". Hours may exceed 24; a negative total keeps its minus sign; seconds appear only when non-zero.

| Parameter | Type | Description |
|---|---|---|
| `expr` | `str` | Terms joined by + and -, e.g. 1:30 + 2:45 - 0:20. |

Returns `str`. Example: `math.time_sum("1:30 + 2:45")`

## math.ratio_simplify

`math.ratio_simplify(ratio: str) -> str`

Reduces a ratio to the smallest whole numbers: "1920:1080" gives "16:9", "4:6:8" gives "2:3:4", "0.5:1.5" gives "1:3". Separators : / x , or spaces.

| Parameter | Type | Description |
|---|---|---|
| `ratio` | `str` | Ratio such as 16:9. |

Returns `str`. Example: `math.ratio_simplify("1920:1080")`

## math.temperature_text

`math.temperature_text(text: str) -> Dict[str, str]`

Converts text like "72f", "-40 C" or "300K" to all four scales and returns a dict with c, f, k and r.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Number with a unit letter, e.g. 72f. |

Returns `Dict[str, str]`. Example: `math.temperature_text("72f")`

## math.what_percent

`math.what_percent(a: float, b: float) -> float`

Returns what percent a is of b (what_percent(20, 50) = 40). b must not be zero.

| Parameter | Type | Description |
|---|---|---|
| `a` | `float` | a value. |
| `b` | `float` | b value. |

Returns `float`. Example: `math.what_percent(20, 50)`

## math.percent_change

`math.percent_change(old: float, new: float) -> float`

Returns the percentage change from old to new ((new-old)/|old|*100; 80 to 100 is 25). old must not be zero.

| Parameter | Type | Description |
|---|---|---|
| `old` | `float` | old value. |
| `new` | `float` | new value. |

Returns `float`. Example: `math.percent_change(80, 100)`

## math.add_percent

`math.add_percent(x: float, p: float) -> float`

Increases x by p percent (add_percent(200, 10) = 220).

| Parameter | Type | Description |
|---|---|---|
| `x` | `float` | x value. |
| `p` | `float` | p value. |

Returns `float`. Example: `math.add_percent(200, 10)`

## math.sub_percent

`math.sub_percent(x: float, p: float) -> float`

Decreases x by p percent (sub_percent(200, 10) = 180).

| Parameter | Type | Description |
|---|---|---|
| `x` | `float` | x value. |
| `p` | `float` | p value. |

Returns `float`. Example: `math.sub_percent(200, 10)`

## math.percent_diff

`math.percent_diff(a: float, b: float) -> float`

Returns the symmetric percentage difference between a and b: |a-b| divided by their average (50 and 100 give 66.667). Both zero gives 0.

| Parameter | Type | Description |
|---|---|---|
| `a` | `float` | a value. |
| `b` | `float` | b value. |

Returns `float`. Example: `math.percent_diff(50, 100)`

## math.markup

`math.markup(cost: float, price: float) -> float`

Returns the markup percentage of selling at price when the cost is cost: (price-cost)/cost*100. cost must not be zero.

| Parameter | Type | Description |
|---|---|---|
| `cost` | `float` | cost value. |
| `price` | `float` | price value. |

Returns `float`. Example: `math.markup(80, 100)`

## math.margin

`math.margin(cost: float, price: float) -> float`

Returns the profit margin percentage of selling at price with cost cost: (price-cost)/price*100. price must not be zero.

| Parameter | Type | Description |
|---|---|---|
| `cost` | `float` | cost value. |
| `price` | `float` | price value. |

Returns `float`. Example: `math.margin(80, 100)`

## math.price_from_markup

`math.price_from_markup(cost: float, markup_pct: float) -> float`

Returns the selling price for a cost and a markup percentage (cost 80, markup 25 gives 100).

| Parameter | Type | Description |
|---|---|---|
| `cost` | `float` | cost value. |
| `markup_pct` | `float` | markup_pct value. |

Returns `float`. Example: `math.price_from_markup(80, 25)`

## math.price_from_margin

`math.price_from_margin(cost: float, margin_pct: float) -> float`

Returns the selling price needed for a target profit margin percentage (cost 80, margin 20 gives 100). margin_pct must be below 100.

| Parameter | Type | Description |
|---|---|---|
| `cost` | `float` | cost value. |
| `margin_pct` | `float` | margin_pct value. |

Returns `float`. Example: `math.price_from_margin(80, 20)`

## math.tax_add

`math.tax_add(amount: float, rate: float) -> float`

Adds tax to a net amount: amount * (1 + rate/100) (tax_add(100, 20) = 120).

| Parameter | Type | Description |
|---|---|---|
| `amount` | `float` | amount value. |
| `rate` | `float` | rate value. |

Returns `float`. Example: `math.tax_add(100, 20)`

## math.tax_remove

`math.tax_remove(amount: float, rate: float) -> float`

Removes tax from a gross amount: amount / (1 + rate/100) (tax_remove(120, 20) = 100). rate must be above -100.

| Parameter | Type | Description |
|---|---|---|
| `amount` | `float` | amount value. |
| `rate` | `float` | rate value. |

Returns `float`. Example: `math.tax_remove(120, 20)`

## math.reverse_percent

`math.reverse_percent(final: float, p: float, increase: bool = True) -> float`

Finds the original value before a percentage was added (increase=true) or taken off (increase=false): reverse_percent(120, 20) = 100, reverse_percent(80, 20, False) = 100.

| Parameter | Type | Description |
|---|---|---|
| `final` | `float` | The value after the change. |
| `p` | `float` | The percentage that was applied. |
| `increase` | `bool` (default `True`) | True if p was added, false if it was subtracted. |

Returns `float`. Example: `math.reverse_percent(120, 20)`

## math.markup_margin

`math.markup_margin(cost: float, price: float) -> Dict[str, str]`

Returns profit, markup percent and margin percent for a cost and a selling price as a dict.

| Parameter | Type | Description |
|---|---|---|
| `cost` | `float` | Cost. |
| `price` | `float` | Selling price. |

Returns `Dict[str, str]`. Example: `math.markup_margin(80, 100)`

## math.discount

`math.discount(price: float, pct: float, pct2: float = 0.0) -> Dict[str, str]`

Applies a percentage discount (optionally a second stacked one) to a price and returns final price, amount saved and the effective percent as a dict.

| Parameter | Type | Description |
|---|---|---|
| `price` | `float` | Original price. |
| `pct` | `float` | Discount percent. |
| `pct2` | `float` (default `0.0`) | Optional second discount percent applied after the first. |

Returns `Dict[str, str]`. Example: `math.discount(80, 25)`

## math.vat

`math.vat(amount: float, rate: float, inclusive: bool = False) -> Dict[str, str]`

Splits an amount into net, VAT and gross. With inclusive=false the amount is net and VAT is added; with inclusive=true the amount already includes VAT and is split out.

| Parameter | Type | Description |
|---|---|---|
| `amount` | `float` | The amount. |
| `rate` | `float` | VAT rate in percent. |
| `inclusive` | `bool` (default `False`) | The amount already includes VAT. |

Returns `Dict[str, str]`. Example: `math.vat(100, 20)`

## math.tip

`math.tip(bill: float, pct: float = 15.0, people: int = 1) -> Dict[str, str]`

Calculates a tip and splits the bill: returns tip, total, per_person, tip_per_person and bill_per_person (rounded to cents) for a bill, a tip percent and a number of people.

| Parameter | Type | Description |
|---|---|---|
| `bill` | `float` | Bill before tip. |
| `pct` | `float` (default `15.0`) | Tip percent. |
| `people` | `int` (default `1`) | Number of people sharing. |

Returns `Dict[str, str]`. Example: `math.tip(85.5, 18, 2)`

## math.sum

`math.sum(values: List[float]) -> float`

Adds up a list of numbers with compensated summation (sum(\[0.1, 0.2, 0.3\]) = 0.6). Empty list is an error.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `float`. Example: `math.sum([1, 2, 3.5])`

## math.mean

`math.mean(values: List[float]) -> float`

Arithmetic mean (average) of a list of numbers.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `float`. Example: `math.mean([2, 4, 9])`

## math.median

`math.median(values: List[float]) -> float`

Median of a list: the middle value, or the mean of the two middle values for an even count.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `float`. Example: `math.median([5, 1, 3, 2])`

## math.mode

`math.mode(values: List[float]) -> List[float]`

Returns the most frequent value(s) in ascending order; several values are returned when they tie (all values if all are unique).

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `List[float]`. Example: `math.mode([1, 2, 2, 3, 3])`

## math.min

`math.min(values: List[float]) -> float`

Smallest value in a list.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `float`. Example: `math.min([4, 2, 9])`

## math.max

`math.max(values: List[float]) -> float`

Largest value in a list.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `float`. Example: `math.max([4, 2, 9])`

## math.range

`math.range(values: List[float]) -> float`

Spread of a list: max minus min.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `float`. Example: `math.range([4, 2, 9])`

## math.count

`math.count(values: List[float]) -> int`

Number of values in a list.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `int`. Example: `math.count([4, 2, 9])`

## math.geometric_mean

`math.geometric_mean(values: List[float]) -> float`

Geometric mean of positive numbers (the n-th root of their product, computed via logs). Zero or negative values are an error.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `float`. Example: `math.geometric_mean([1, 4, 16])`

## math.harmonic_mean

`math.harmonic_mean(values: List[float]) -> float`

Harmonic mean of positive numbers (count divided by the sum of reciprocals), suited to averaging rates. Zero or negative values are an error.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `float`. Example: `math.harmonic_mean([40, 60])`

## math.iqr

`math.iqr(values: List[float]) -> float`

Interquartile range: the 75th minus the 25th percentile (linear interpolation).

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `float`. Example: `math.iqr([1, 2, 3, 4, 5, 6, 7, 8])`

## math.normalize

`math.normalize(values: List[float]) -> List[float]`

Rescales values linearly to the range 0..1 (min-max normalisation). If all values are equal every result is 0.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `List[float]`. Example: `math.normalize([10, 20, 30])`

## math.cumulative_sum

`math.cumulative_sum(values: List[float]) -> List[float]`

Running total: element i is the sum of the first i+1 values.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |

Returns `List[float]`. Example: `math.cumulative_sum([1, 2, 3, 4])`

## math.variance

`math.variance(values: List[float], sample: bool = True) -> float`

Variance of a list: sample variance (n-1 divisor, default) or population variance with sample=false. Sample variance needs at least 2 values.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |
| `sample` | `bool` (default `True`) | Use the sample (n-1) formula; false gives population variance. |

Returns `float`. Example: `math.variance([2, 4, 4, 4, 5, 5, 7, 9], sample=False)`

## math.stdev

`math.stdev(values: List[float], sample: bool = True) -> float`

Standard deviation: sample (n-1 divisor, default) or population with sample=false. For \[2,4,4,4,5,5,7,9\] the population value is 2.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |
| `sample` | `bool` (default `True`) | Use the sample (n-1) formula; false gives population stdev. |

Returns `float`. Example: `math.stdev([2, 4, 4, 4, 5, 5, 7, 9], sample=False)`

## math.zscores

`math.zscores(values: List[float], sample: bool = False) -> List[float]`

Standardises values: (x - mean) / stdev for each value. Uses the population stdev unless sample=true. All-equal data is an error (stdev is 0).

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |
| `sample` | `bool` (default `False`) | Use the sample stdev. |

Returns `List[float]`. Example: `math.zscores([1, 2, 3, 4, 5])`

## math.quantile

`math.quantile(values: List[float], q: float) -> float`

Quantile q (0..1) with linear interpolation between ranks (NumPy default): quantile(v, 0.5) is the median.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |
| `q` | `float` | Fraction between 0 and 1. |

Returns `float`. Example: `math.quantile([1, 2, 3, 4], 0.25)`

## math.percentile

`math.percentile(values: List[float], p: float) -> float`

Percentile p (0..100) with linear interpolation: percentile(v, 90) is the value below which 90 percent of the data fall.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |
| `p` | `float` | Percent between 0 and 100. |

Returns `float`. Example: `math.percentile([1, 2, 3, 4, 5], 90)`

## math.moving_average

`math.moving_average(values: List[float], window: int = 3) -> List[float]`

Simple moving average with the given window; returns len(values)-window+1 averages (only full windows). window must be 1..len(values).

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |
| `window` | `int` (default `3`) | Window size. |

Returns `List[float]`. Example: `math.moving_average([1, 2, 3, 4, 5], 3)`

## math.outliers

`math.outliers(values: List[float], k: float = 1.5) -> List[float]`

Returns the values outside the Tukey fences Q1 - k*IQR and Q3 + k*IQR (k defaults to 1.5), in input order.

| Parameter | Type | Description |
|---|---|---|
| `values` | `List[float]` | The numbers (a non-empty list of floats). |
| `k` | `float` (default `1.5`) | Fence multiplier. |

Returns `List[float]`. Example: `math.outliers([1, 2, 2, 3, 3, 4, 100])`

## math.correlation

`math.correlation(xs: List[float], ys: List[float]) -> float`

Pearson correlation coefficient of two equal-length lists (-1..1). Needs at least 2 points and non-constant data.

| Parameter | Type | Description |
|---|---|---|
| `xs` | `List[float]` | X values. |
| `ys` | `List[float]` | Y values. |

Returns `float`. Example: `math.correlation([1, 2, 3], [2, 4, 6])`

## math.linear_regression

`math.linear_regression(xs: List[float], ys: List[float]) -> Dict[str, str]`

Least-squares line y = slope*x + intercept through the points; returns slope, intercept, r and r2 as a dict. xs must not all be equal.

| Parameter | Type | Description |
|---|---|---|
| `xs` | `List[float]` | X values. |
| `ys` | `List[float]` | Y values. |

Returns `Dict[str, str]`. Example: `math.linear_regression([1, 2, 3], [2, 4, 6])`

## math.round_to

`math.round_to(x: float, decimals: int = 0, mode: str = "half_up") -> float`

Rounds x to a number of decimals (negative rounds to tens, hundreds) using its decimal form, so 2.675 gives 2.68 with half_up. Modes: half_up (default, away from zero on ties), half_even (banker's), half_down, floor, ceil, trunc.

| Parameter | Type | Description |
|---|---|---|
| `x` | `float` | Number to round. |
| `decimals` | `int` (default `0`) | Decimal places (-300..300). |
| `mode` | `str` (default `"half_up"`) | Rounding mode. One of: half_up, half_even, half_down, floor, ceil, trunc. |

Returns `float`. Example: `math.round_to(2.675, 2)`

## math.round_to_multiple

`math.round_to_multiple(x: float, multiple: float, mode: str = "half_up") -> float`

Rounds x to the nearest multiple of m (round_to_multiple(17, 5) = 15, (18, 5) = 20, 0.1 steps work exactly). Same modes as round_to. m must not be zero.

| Parameter | Type | Description |
|---|---|---|
| `x` | `float` | Number to round. |
| `multiple` | `float` | The step. |
| `mode` | `str` (default `"half_up"`) | Rounding mode. One of: half_up, half_even, half_down, floor, ceil, trunc. |

Returns `float`. Example: `math.round_to_multiple(18, 5)`

## math.clamp

`math.clamp(x: float, lo: float, hi: float) -> float`

Limits x to the range lo..hi (clamp(15, 0, 10) = 10). lo must not exceed hi.

| Parameter | Type | Description |
|---|---|---|
| `x` | `float` | Value. |
| `lo` | `float` | Lower bound. |
| `hi` | `float` | Upper bound. |

Returns `float`. Example: `math.clamp(15, 0, 10)`

## math.lerp

`math.lerp(a: float, b: float, t: float) -> float`

Linear interpolation a + (b-a)*t (lerp(10, 20, 0.25) = 12.5); t outside 0..1 extrapolates.

| Parameter | Type | Description |
|---|---|---|
| `a` | `float` | Start. |
| `b` | `float` | End. |
| `t` | `float` | Position, 0 = a, 1 = b. |

Returns `float`. Example: `math.lerp(10, 20, 0.25)`

## math.inverse_lerp

`math.inverse_lerp(a: float, b: float, x: float) -> float`

Returns where x sits between a and b as a fraction ((x-a)/(b-a)). a and b must differ.

| Parameter | Type | Description |
|---|---|---|
| `a` | `float` | Start. |
| `b` | `float` | End. |
| `x` | `float` | Value. |

Returns `float`. Example: `math.inverse_lerp(10, 20, 12.5)`

## math.map_range

`math.map_range(x: float, in_lo: float, in_hi: float, out_lo: float, out_hi: float) -> float`

Maps x from the range in_lo..in_hi to out_lo..out_hi linearly (map_range(5, 0, 10, 0, 100) = 50). in_lo and in_hi must differ; values outside are extrapolated.

| Parameter | Type | Description |
|---|---|---|
| `x` | `float` | Value. |
| `in_lo` | `float` | Input range start. |
| `in_hi` | `float` | Input range end. |
| `out_lo` | `float` | Output range start. |
| `out_hi` | `float` | Output range end. |

Returns `float`. Example: `math.map_range(5, 0, 10, 0, 100)`

## math.solve_linear

`math.solve_linear(a: float, b: float) -> float`

Solves a*x + b = 0 for x (solve_linear(2, -6) = 3). a must not be zero.

| Parameter | Type | Description |
|---|---|---|
| `a` | `float` | Coefficient of x. |
| `b` | `float` | Constant term. |

Returns `float`. Example: `math.solve_linear(2, -6)`

## math.solve_linear_system_2x2

`math.solve_linear_system_2x2(matrix: List[float]) -> List[float]`

Solves two linear equations in x and y given as the augmented matrix \[a1, b1, c1, a2, b2, c2\] meaning a1*x+b1*y=c1 and a2*x+b2*y=c2. Returns \[x, y\]; singular systems are an error.

| Parameter | Type | Description |
|---|---|---|
| `matrix` | `List[float]` | Six numbers: a1, b1, c1, a2, b2, c2. |

Returns `List[float]`. Example: `math.solve_linear_system_2x2([1, 1, 3, 1, -1, 1])`

## math.solve_linear_system_3x3

`math.solve_linear_system_3x3(matrix: List[float]) -> List[float]`

Solves three linear equations in x, y, z given as the row-by-row augmented matrix (12 numbers: a, b, c, d for each equation a*x+b*y+c*z=d). Returns \[x, y, z\]; singular systems are an error.

| Parameter | Type | Description |
|---|---|---|
| `matrix` | `List[float]` | Twelve numbers, four per equation. |

Returns `List[float]`. Example: `math.solve_linear_system_3x3([1,1,1,6, 0,2,5,-4, 2,5,-1,27])`

## math.simplify_fraction

`math.simplify_fraction(n: int, d: int) -> str`

Reduces n/d to lowest terms with a positive denominator (simplify_fraction(6, -8) = "-3/4"; whole results print without a denominator). d must not be zero.

| Parameter | Type | Description |
|---|---|---|
| `n` | `int` | Numerator. |
| `d` | `int` | Denominator. |

Returns `str`. Example: `math.simplify_fraction(6, 8)`

## math.fraction_add

`math.fraction_add(a: str, b: str) -> str`

Exact fraction addition of a and b (each "3/4", "0.75", "2" or "1 1/2"). The result is in lowest terms, printed as n/d or a whole number.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First fraction. |
| `b` | `str` | Second fraction. |

Returns `str`. Example: `math.fraction_add("1/2", "1/3")`

## math.fraction_div

`math.fraction_div(a: str, b: str) -> str`

Exact fraction division of a and b (each "3/4", "0.75", "2" or "1 1/2"). The result is in lowest terms, printed as n/d or a whole number.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First fraction. |
| `b` | `str` | Second fraction. |

Returns `str`. Example: `math.fraction_div("1/2", "1/3")`

## math.fraction_mul

`math.fraction_mul(a: str, b: str) -> str`

Exact fraction multiplication of a and b (each "3/4", "0.75", "2" or "1 1/2"). The result is in lowest terms, printed as n/d or a whole number.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First fraction. |
| `b` | `str` | Second fraction. |

Returns `str`. Example: `math.fraction_mul("1/2", "1/3")`

## math.fraction_sub

`math.fraction_sub(a: str, b: str) -> str`

Exact fraction subtraction of a and b (each "3/4", "0.75", "2" or "1 1/2"). The result is in lowest terms, printed as n/d or a whole number.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First fraction. |
| `b` | `str` | Second fraction. |

Returns `str`. Example: `math.fraction_sub("1/2", "1/3")`

## math.decimal_to_fraction

`math.decimal_to_fraction(x: str, max_denominator: int = 0) -> str`

Converts a decimal to a fraction. With max_denominator=0 (default) the conversion is exact for the decimal text (0.125 = 1/8, 0.333 = 333/1000); a positive max_denominator finds the closest fraction up to that denominator (3.14159 with 1000 = 355/113).

| Parameter | Type | Description |
|---|---|---|
| `x` | `str` | Decimal number or fraction. |
| `max_denominator` | `int` (default `0`) | Largest denominator (0 = exact). |

Returns `str`. Example: `math.decimal_to_fraction("0.125")`

## math.mixed_number

`math.mixed_number(x: str) -> str`

Writes an improper fraction or decimal as a mixed number ("7/2" gives "3 1/2", 2.75 gives "2 3/4", -7/2 gives "-3 1/2").

| Parameter | Type | Description |
|---|---|---|
| `x` | `str` | Fraction like 7/2 or a decimal. |

Returns `str`. Example: `math.mixed_number("7/2")`

## math.scale_ratio

`math.scale_ratio(ratio: str, known: float, position: int = 1) -> str`

Scales a ratio so that one part has a known value and returns all parts: scale_ratio("16:9", 1920) gives "1920:1080"; position picks which part (1-based) the value belongs to.

| Parameter | Type | Description |
|---|---|---|
| `ratio` | `str` | Ratio such as 16:9. |
| `known` | `float` | Value of the chosen part. |
| `position` | `int` (default `1`) | 1-based index of the part that equals known. |

Returns `str`. Example: `math.scale_ratio("16:9", 1920)`

## math.temperature

`math.temperature(value: float, from_unit: str, to_unit: str) -> float`

Converts a temperature between c (Celsius), f (Fahrenheit), k (Kelvin) and r (Rankine); full unit names and a degree sign are accepted. Values below absolute zero are an error.

| Parameter | Type | Description |
|---|---|---|
| `value` | `float` | Temperature. |
| `from_unit` | `str` | Source unit: c, f, k or r. |
| `to_unit` | `str` | Target unit: c, f, k or r. |

Returns `float`. Example: `math.temperature(100, "c", "f")`

## math.degrees_to_radians

`math.degrees_to_radians(degrees: float) -> float`

Converts degrees to radians (180 gives 3.14159...).

| Parameter | Type | Description |
|---|---|---|
| `degrees` | `float` | Angle in degrees. |

Returns `float`. Example: `math.degrees_to_radians(180)`

## math.radians_to_degrees

`math.radians_to_degrees(radians: float) -> float`

Converts radians to degrees (pi gives 180).

| Parameter | Type | Description |
|---|---|---|
| `radians` | `float` | Angle in radians. |

Returns `float`. Example: `math.radians_to_degrees(3.141592653589793)`

## math.polar_to_cartesian

`math.polar_to_cartesian(r: float, theta: float, degrees: bool = False) -> Dict[str, str]`

Converts polar (r, theta) to Cartesian x and y. theta is in radians unless degrees=true. Returns a dict with x and y (tiny floating-point noise below 1e-12 is cleaned to 0).

| Parameter | Type | Description |
|---|---|---|
| `r` | `float` | Radius. |
| `theta` | `float` | Angle. |
| `degrees` | `bool` (default `False`) | theta is in degrees. |

Returns `Dict[str, str]`. Example: `math.polar_to_cartesian(2, 90, degrees=True)`

## math.cartesian_to_polar

`math.cartesian_to_polar(x: float, y: float, degrees: bool = False) -> Dict[str, str]`

Converts Cartesian (x, y) to polar: returns r and theta (radians, or degrees with degrees=true, in the range -pi..pi / -180..180).

| Parameter | Type | Description |
|---|---|---|
| `x` | `float` | X. |
| `y` | `float` | Y. |
| `degrees` | `bool` (default `False`) | Return theta in degrees. |

Returns `Dict[str, str]`. Example: `math.cartesian_to_polar(1, 1, degrees=True)`

## math.distance_2d

`math.distance_2d(x1: float, y1: float, x2: float, y2: float) -> float`

Euclidean distance between (x1, y1) and (x2, y2).

| Parameter | Type | Description |
|---|---|---|
| `x1` | `float` | X1. |
| `y1` | `float` | Y1. |
| `x2` | `float` | X2. |
| `y2` | `float` | Y2. |

Returns `float`. Example: `math.distance_2d(0, 0, 3, 4)`

## math.distance_3d

`math.distance_3d(x1: float, y1: float, z1: float, x2: float, y2: float, z2: float) -> float`

Euclidean distance between (x1, y1, z1) and (x2, y2, z2).

| Parameter | Type | Description |
|---|---|---|
| `x1` | `float` | X1. |
| `y1` | `float` | Y1. |
| `z1` | `float` | Z1. |
| `x2` | `float` | X2. |
| `y2` | `float` | Y2. |
| `z2` | `float` | Z2. |

Returns `float`. Example: `math.distance_3d(0, 0, 0, 1, 2, 2)`

## math.haversine

`math.haversine(lat1: float, lon1: float, lat2: float, lon2: float) -> Dict[str, str]`

Great-circle distance between two latitude/longitude points (degrees) using the haversine formula on a 6371.0088 km sphere. Returns km, mi, nmi and the initial bearing in degrees. Latitudes must be -90..90 and longitudes -180..180.

| Parameter | Type | Description |
|---|---|---|
| `lat1` | `float` | Latitude 1. |
| `lon1` | `float` | Longitude 1. |
| `lat2` | `float` | Latitude 2. |
| `lon2` | `float` | Longitude 2. |

Returns `Dict[str, str]`. Example: `math.haversine(51.5074, -0.1278, 48.8566, 2.3522)`

## math.geometry

`math.geometry(shape: str, a: float = 0.0, b: float = 0.0, c: float = 0.0) -> Dict[str, str]`

Area, perimeter, volume and other measures of a shape as a dict. circle(a=radius), rectangle(a=width, b=height), square(a), triangle(a, b, c = three sides), ellipse(a, b = semi-axes), sphere(a=radius), cube(a), cylinder(a=radius, b=height), cone(a=radius, b=height). Unused dimensions stay 0; non-positive needed ones are an error.

| Parameter | Type | Description |
|---|---|---|
| `shape` | `str` | Shape name. |
| `a` | `float` (default `0.0`) | First dimension. |
| `b` | `float` (default `0.0`) | Second dimension. |
| `c` | `float` (default `0.0`) | Third dimension. |

Returns `Dict[str, str]`. Example: `math.geometry("cylinder", 2, 5)`

## math.circle_area

`math.circle_area(r: float) -> float`

Area of a circle: pi*r^2.

| Parameter | Type | Description |
|---|---|---|
| `r` | `float` | Radius. |

Returns `float`. Example: `math.circle_area(3)`

## math.circle_circumference

`math.circle_circumference(r: float) -> float`

Circumference of a circle: 2*pi*r.

| Parameter | Type | Description |
|---|---|---|
| `r` | `float` | Radius. |

Returns `float`. Example: `math.circle_circumference(3)`

## math.rect_area

`math.rect_area(w: float, h: float) -> float`

Area of a rectangle: w*h.

| Parameter | Type | Description |
|---|---|---|
| `w` | `float` | Width. |
| `h` | `float` | Height. |

Returns `float`. Example: `math.rect_area(3, 4)`

## math.rect_perimeter

`math.rect_perimeter(w: float, h: float) -> float`

Perimeter of a rectangle: 2*(w+h).

| Parameter | Type | Description |
|---|---|---|
| `w` | `float` | Width. |
| `h` | `float` | Height. |

Returns `float`. Example: `math.rect_perimeter(3, 4)`

## math.triangle_area

`math.triangle_area(a: float, b: float, c: float) -> float`

Area of a triangle from its three side lengths (Heron's formula; 3, 4, 5 gives 6). Sides that violate the triangle inequality are an error.

| Parameter | Type | Description |
|---|---|---|
| `a` | `float` | Side a. |
| `b` | `float` | Side b. |
| `c` | `float` | Side c. |

Returns `float`. Example: `math.triangle_area(3, 4, 5)`

## math.triangle_perimeter

`math.triangle_perimeter(a: float, b: float, c: float) -> float`

Perimeter of a triangle: a+b+c (sides must satisfy the triangle inequality).

| Parameter | Type | Description |
|---|---|---|
| `a` | `float` | Side a. |
| `b` | `float` | Side b. |
| `c` | `float` | Side c. |

Returns `float`. Example: `math.triangle_perimeter(3, 4, 5)`

## math.sphere_volume

`math.sphere_volume(r: float) -> float`

Volume of a sphere: 4/3*pi*r^3.

| Parameter | Type | Description |
|---|---|---|
| `r` | `float` | Radius. |

Returns `float`. Example: `math.sphere_volume(2)`

## math.sphere_area

`math.sphere_area(r: float) -> float`

Surface area of a sphere: 4*pi*r^2.

| Parameter | Type | Description |
|---|---|---|
| `r` | `float` | Radius. |

Returns `float`. Example: `math.sphere_area(2)`

## math.cylinder_volume

`math.cylinder_volume(r: float, h: float) -> float`

Volume of a cylinder: pi*r^2*h.

| Parameter | Type | Description |
|---|---|---|
| `r` | `float` | Radius. |
| `h` | `float` | Height. |

Returns `float`. Example: `math.cylinder_volume(2, 5)`

## math.cylinder_area

`math.cylinder_area(r: float, h: float) -> float`

Total surface area of a cylinder: 2*pi*r*(r+h).

| Parameter | Type | Description |
|---|---|---|
| `r` | `float` | Radius. |
| `h` | `float` | Height. |

Returns `float`. Example: `math.cylinder_area(2, 5)`

## math.cone_volume

`math.cone_volume(r: float, h: float) -> float`

Volume of a cone: pi*r^2*h/3.

| Parameter | Type | Description |
|---|---|---|
| `r` | `float` | Radius. |
| `h` | `float` | Height. |

Returns `float`. Example: `math.cone_volume(2, 5)`

## math.cone_area

`math.cone_area(r: float, h: float) -> float`

Total surface area of a cone: pi*r*(r+slant) with slant = sqrt(r^2+h^2).

| Parameter | Type | Description |
|---|---|---|
| `r` | `float` | Radius. |
| `h` | `float` | Height. |

Returns `float`. Example: `math.cone_area(2, 5)`

## math.pythagoras

`math.pythagoras(a: float = 0.0, b: float = 0.0, c: float = 0.0) -> float`

Solves a right triangle with legs a, b and hypotenuse c: set exactly one of them to 0 (the unknown) and the other two to their values. pythagoras(3, 4, 0) = 5; pythagoras(0, 12, 13) = 5. The hypotenuse must be longer than a leg.

| Parameter | Type | Description |
|---|---|---|
| `a` | `float` (default `0.0`) | Leg a (0 if unknown). |
| `b` | `float` (default `0.0`) | Leg b (0 if unknown). |
| `c` | `float` (default `0.0`) | Hypotenuse c (0 if unknown). |

Returns `float`. Example: `math.pythagoras(3, 4, 0)`

## math.compound_interest

`math.compound_interest(principal: float, rate: float, years: float, n: int = 1) -> Dict[str, str]`

Future value with compound interest: principal * (1 + rate/n)^(n*years), rate in percent per year, n compounding periods per year (n=1 annual, 12 monthly, 0 continuous). Returns amount, interest and the effective annual rate as a dict.

| Parameter | Type | Description |
|---|---|---|
| `principal` | `float` | Starting amount. |
| `rate` | `float` | Annual rate in percent. |
| `years` | `float` | Duration in years. |
| `n` | `int` (default `1`) | Compounding periods per year (0 = continuous). |

Returns `Dict[str, str]`. Example: `math.compound_interest(1000, 5, 10, 1)`

## math.simple_interest

`math.simple_interest(principal: float, rate: float, years: float) -> Dict[str, str]`

Simple (non-compounding) interest: principal * rate/100 * years. Returns interest and total as a dict.

| Parameter | Type | Description |
|---|---|---|
| `principal` | `float` | Amount. |
| `rate` | `float` | Annual rate in percent. |
| `years` | `float` | Duration in years. |

Returns `Dict[str, str]`. Example: `math.simple_interest(1000, 5, 3)`

## math.amortization_summary

`math.amortization_summary(principal: float, annual_rate: float, years: float, per_year: int = 12, after: int = 12) -> Dict[str, str]`

Splits a loan into interest and principal: after `after` payments returns interest_paid, principal_paid, balance, first_interest, first_principal, payment and payments_left (same inputs as loan_payment).

| Parameter | Type | Description |
|---|---|---|
| `principal` | `float` | Amount borrowed. |
| `annual_rate` | `float` | Annual rate in percent. |
| `years` | `float` | Term in years. |
| `per_year` | `int` (default `12`) | Payments per year. |
| `after` | `int` (default `12`) | Number of payments already made. |

Returns `Dict[str, str]`. Example: `math.amortization_summary(200000, 6, 30, 12, 12)`

## math.future_value

`math.future_value(present: float, rate: float, periods: float, payment: float = 0.0) -> float`

Future value of a starting amount plus optional regular payments at the end of each period: present*(1+r)^n + payment*((1+r)^n-1)/r, with rate the percent per period.

| Parameter | Type | Description |
|---|---|---|
| `present` | `float` | Starting amount. |
| `rate` | `float` | Rate per period in percent. |
| `periods` | `float` | Number of periods. |
| `payment` | `float` (default `0.0`) | Payment added at the end of every period. |

Returns `float`. Example: `math.future_value(1000, 5, 10, 100)`

## math.present_value

`math.present_value(future: float, rate: float, periods: float, payment: float = 0.0) -> float`

Present value of a future amount plus optional regular payments, discounted at a percent rate per period: future/(1+r)^n + payment*(1-(1+r)^-n)/r.

| Parameter | Type | Description |
|---|---|---|
| `future` | `float` | Future amount. |
| `rate` | `float` | Discount rate per period in percent. |
| `periods` | `float` | Number of periods. |
| `payment` | `float` (default `0.0`) | Payment received at the end of every period. |

Returns `float`. Example: `math.present_value(1628.89, 5, 10)`

## math.cagr

`math.cagr(start: float, end: float, years: float) -> float`

Compound annual growth rate in percent from a start value to an end value over some years ((end/start)^(1/years)-1). Start and end must be positive; years &gt; 0.

| Parameter | Type | Description |
|---|---|---|
| `start` | `float` | Start value. |
| `end` | `float` | End value. |
| `years` | `float` | Years. |

Returns `float`. Example: `math.cagr(1000, 2000, 5)`

## math.roi

`math.roi(initial: float, final: float) -> float`

Return on investment in percent: (final - initial) / initial * 100. initial must not be zero.

| Parameter | Type | Description |
|---|---|---|
| `initial` | `float` | Amount invested. |
| `final` | `float` | Value at the end. |

Returns `float`. Example: `math.roi(1000, 1250)`

## math.rule_of_72

`math.rule_of_72(rate: float) -> Dict[str, str]`

Years to double money at an annual percent rate: returns the rule-of-72 estimate and the exact value ln(2)/ln(1+r) as a dict. rate must be positive.

| Parameter | Type | Description |
|---|---|---|
| `rate` | `float` | Annual rate in percent. |

Returns `Dict[str, str]`. Example: `math.rule_of_72(6)`

## math.break_even

`math.break_even(fixed_costs: float, price: float, variable_cost: float) -> Dict[str, str]`

Break-even point: units = fixed_costs / (price - variable_cost). Returns exact units, units rounded up and the revenue at break-even. price must exceed variable_cost.

| Parameter | Type | Description |
|---|---|---|
| `fixed_costs` | `float` | Fixed costs. |
| `price` | `float` | Selling price per unit. |
| `variable_cost` | `float` | Variable cost per unit. |

Returns `Dict[str, str]`. Example: `math.break_even(5000, 25, 15)`

## math.unit_price

`math.unit_price(price: float, qty: float) -> float`

Price per unit: price / quantity (quantity must be positive).

| Parameter | Type | Description |
|---|---|---|
| `price` | `float` | Total price. |
| `qty` | `float` | Quantity. |

Returns `float`. Example: `math.unit_price(3.5, 12)`

## math.compare_unit_price

`math.compare_unit_price(a_price: float, a_qty: float, b_price: float, b_qty: float) -> Dict[str, str]`

Compares two offers by unit price: returns each unit price, which is cheaper (a, b or same) and the saving percent of the cheaper one.

| Parameter | Type | Description |
|---|---|---|
| `a_price` | `float` | Price of offer A. |
| `a_qty` | `float` | Quantity of offer A. |
| `b_price` | `float` | Price of offer B. |
| `b_qty` | `float` | Quantity of offer B. |

Returns `Dict[str, str]`. Example: `math.compare_unit_price(3.5, 12, 5, 20)`

## math.average_speed

`math.average_speed(distance: float, time: float) -> float`

Average speed = distance / time (any consistent units). time must be positive.

| Parameter | Type | Description |
|---|---|---|
| `distance` | `float` | distance. |
| `time` | `float` | time. |

Returns `float`. Example: `math.average_speed(150, 2.5)`

## math.travel_time

`math.travel_time(distance: float, speed: float) -> float`

Time = distance / speed. speed must be positive.

| Parameter | Type | Description |
|---|---|---|
| `distance` | `float` | distance. |
| `speed` | `float` | speed. |

Returns `float`. Example: `math.travel_time(150, 60)`

## math.travel_distance

`math.travel_distance(speed: float, time: float) -> float`

Distance = speed * time.

| Parameter | Type | Description |
|---|---|---|
| `speed` | `float` | Speed. |
| `time` | `float` | Time. |

Returns `float`. Example: `math.travel_distance(60, 2.5)`

## math.bmi

`math.bmi(kg: float, m: float) -> Dict[str, str]`

Body mass index = kg / m^2, with the WHO category (underweight, normal, overweight, obese). Weight in kilograms, height in metres; both must be positive.

| Parameter | Type | Description |
|---|---|---|
| `kg` | `float` | Weight in kilograms. |
| `m` | `float` | Height in metres. |

Returns `Dict[str, str]`. Example: `math.bmi(70, 1.75)`

## math.angle_sum

`math.angle_sum(expr: str) -> str`

Adds and subtracts angles written as degrees/minutes/seconds: "30°15'20\" + 45°50'" gives 76°05'20". Each term accepts 40°26'46"N style, 40 26 46 or plain degrees. Result is in d°mm'ss" form.

| Parameter | Type | Description |
|---|---|---|
| `expr` | `str` | Angles joined by + and -. |

Returns `str`. Example: `math.angle_sum("30 15 20 + 45 50")`

## math.dms_to_decimal

`math.dms_to_decimal(dms: str) -> float`

Converts degrees/minutes/seconds (40°26'46"N, 40 26 46, -73°59'W) to decimal degrees; S and W (or a leading minus) make it negative.

| Parameter | Type | Description |
|---|---|---|
| `dms` | `str` | Angle text. |

Returns `float`. Example: `math.dms_to_decimal("40°26'46\"N")`

## math.decimal_to_dms

`math.decimal_to_dms(degrees: float) -> str`

Converts decimal degrees to d°mm'ss.ss" form (40.446 gives 40°26'45.6"); the sign is kept as a minus.

| Parameter | Type | Description |
|---|---|---|
| `degrees` | `float` | Decimal degrees. |

Returns `str`. Example: `math.decimal_to_dms(40.446)`

## math.combinations

`math.combinations(n: int, k: int) -> str`

Number of ways to choose k items from n without order, C(n, k), as a decimal string (combinations(52, 5) = 2598960). k &gt; n gives 0. n is limited to 0..20000.

| Parameter | Type | Description |
|---|---|---|
| `n` | `int` | Total items. |
| `k` | `int` | Items chosen. |

Returns `str`. Example: `math.combinations(52, 5)`

## math.permutations

`math.permutations(n: int, k: int) -> str`

Number of ordered arrangements of k items from n, P(n, k) = n!/(n-k)!, as a decimal string (permutations(10, 3) = 720). k &gt; n gives 0. n is limited to 0..20000.

| Parameter | Type | Description |
|---|---|---|
| `n` | `int` | Total items. |
| `k` | `int` | Items arranged. |

Returns `str`. Example: `math.permutations(10, 3)`

## math.binomial_probability

`math.binomial_probability(n: int, k: int, p: float) -> float`

Probability of exactly k successes in n independent trials with success probability p: C(n,k) p^k (1-p)^(n-k). n up to 10 million; p in 0..1.

| Parameter | Type | Description |
|---|---|---|
| `n` | `int` | Trials. |
| `k` | `int` | Successes. |
| `p` | `float` | Success probability, 0..1. |

Returns `float`. Example: `math.binomial_probability(10, 5, 0.5)`

## math.normal_cdf

`math.normal_cdf(x: float, mean: float = 0.0, sd: float = 1.0) -> float`

Cumulative probability P(X &lt;= x) of a normal distribution with the given mean (default 0) and standard deviation sd (default 1, must be positive). normal_cdf(1.96) is about 0.975.

| Parameter | Type | Description |
|---|---|---|
| `x` | `float` | Value. |
| `mean` | `float` (default `0.0`) | Mean. |
| `sd` | `float` (default `1.0`) | Standard deviation. |

Returns `float`. Example: `math.normal_cdf(1.96)`

## math.normal_pdf

`math.normal_pdf(x: float, mean: float = 0.0, sd: float = 1.0) -> float`

Probability density of a normal distribution at x (mean default 0, sd default 1, sd must be positive).

| Parameter | Type | Description |
|---|---|---|
| `x` | `float` | Value. |
| `mean` | `float` (default `0.0`) | Mean. |
| `sd` | `float` (default `1.0`) | Standard deviation. |

Returns `float`. Example: `math.normal_pdf(0)`

## math.dice_probability

`math.dice_probability(sides: int, count: int, target: int) -> float`

Probability that `count` dice with `sides` sides sum to exactly `target` (two six-sided dice summing to 7 gives 1/6). count*sides is limited to 5000.

| Parameter | Type | Description |
|---|---|---|
| `sides` | `int` | Sides per die (>= 2). |
| `count` | `int` | Number of dice (>= 1). |
| `target` | `int` | Target sum. |

Returns `float`. Example: `math.dice_probability(6, 2, 7)`

