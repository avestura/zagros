"""Math and statistics: An expression evaluator, percentages, statistics, algebra, geometry and finance helpers.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `math` is predeclared in main.star; call `math.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: calculator, expression, evaluate, percentage, statistics, mean, median, stdev, algebra, geometry, area, volume, finance, loan, interest, discount, tip, fraction, unit price
"""

from typing import Any, Dict, List

def eval_str(expr: str, degrees: bool = False) -> str:
    """Evaluates a math expression (same syntax as eval) and returns a clean string with up to 12 significant digits and no trailing zeros, so 0.1+0.2 gives "0.3" and 1/3 gives "0.333333333333". Very large or small results use exponent form (1e+25).

        expr: The expression, e.g. 2*(3+4)^2, sqrt(16), 200 + 10%, 15% of 80.
        degrees: Trig functions take and return degrees instead of radians.

    Example: math.eval_str("0.1 + 0.2")
    Keywords: calculator, calc, evaluate, expression, arithmetic, formula, compute, string, result, clean
    Needs: none (pure function).
    """

def eval(expr: str, degrees: bool = False) -> float:
    """Evaluates a math expression with floats and returns the number. Supports + - * / % (remainder) ^ or ** (power), parentheses, implicit multiplication (2(3+4), 3pi), constants pi e tau phi, factorial 5!, thousands separators (1,000), scientific notation (1e3), functions sin cos tan asin acos atan atan2 sinh cosh tanh sqrt cbrt exp ln log(x) [base 10] log(x,base) log2 log10 abs floor ceil round(x[,d]) trunc sign min max sum avg pow hypot gamma factorial deg rad, and the word 'of' as multiplication. Percent: 50% is 0.5, and in a sum 'a + b%' means a plus b percent of a (200 + 10% = 220); a comma is never a decimal point. Errors name the character position.

        expr: The expression, e.g. 2*(3+4)^2, sqrt(16), 200 + 10%, 15% of 80.
        degrees: Trig functions take and return degrees instead of radians.

    Example: math.eval("2(3+4) + 10%")
    Task: evaluate a math expression
    Keywords: calculator, calculate, compute, arithmetic, formula, calc, evaluate, expression, number, sqrt, sin, cos, log, percent
    Needs: none (pure function).
    """

def percent_text(text: str) -> Dict[str, str]:
    """Answers a percentage question typed in words. Understands "15% of 80", "20 is what % of 50", "what percent is 20 of 50", "50 is 25% of what", "from 80 to 100" (percent change) and "200 + 10%" / "200 - 10%". Returns a dict with result plus an explanation.

        text: The question, e.g. 15% of 80.

    Example: math.percent_text("15% of 80")
    Keywords: percent, percentage, of, change, increase, decrease, what, calculator, ratio
    Needs: none (pure function).
    """

def percent_of(p: float, x: float) -> float:
    """Returns p percent of x (percent_of(15, 80) = 12).

        p: p value.
        x: x value.

    Example: math.percent_of(15, 80)
    Task: take a percentage of a number
    Keywords: what is 15% of, percentage, percent, of, portion, share, part
    Needs: none (pure function).
    """

def stats_text(text: str) -> Dict[str, str]:
    """Extracts every number from any text (commas, newlines or words in between; 1,234 is one number) and returns summary statistics like describe. Errors if the text has no numbers.

        text: Text containing numbers.

    Example: math.stats_text("scores: 3, 5, 8 and 13")
    Keywords: stats, statistics, text, numbers, extract, average, sum, mean, median, describe, paste, list
    Needs: none (pure function).
    """

def describe(values: List[float]) -> Dict[str, str]:
    """Summary statistics as a dict: count, sum, mean, median, mode, min, max, range, sample variance and stdev, pop_variance, pop_stdev, q1, q3, iqr and (for positive data) geometric_mean and harmonic_mean.

        values: The numbers (a non-empty list of floats).

    Example: math.describe([2, 4, 4, 4, 5, 5, 7, 9])
    Keywords: describe, summary, statistics, stats, mean, median, stdev, variance, quartiles
    Needs: none (pure function).
    """

def loan_payment(principal: float, annual_rate: float, years: float, per_year: int = 12) -> Dict[str, str]:
    """Fixed payment of an amortising loan or mortgage: principal, annual rate in percent, term in years and payments per year (default 12). Returns payment, total repaid, total interest and the number of payments, rounded to cents (200000 at 6% for 30 years gives 1199.10 a month).

        principal: Amount borrowed.
        annual_rate: Annual interest rate in percent (0 allowed).
        years: Term in years.
        per_year: Payments per year.

    Example: math.loan_payment(200000, 6, 30)
    Task: monthly loan payment
    Keywords: mortgage payment, emi, installment, loan, payment, mortgage, monthly, amortization, interest, car
    Needs: none (pure function).
    """

def tip_text(text: str) -> Dict[str, str]:
    """Reads "bill [tip%] [people]" from text ("85.50 18% 2" or "85.5 18 2"; tip defaults to 15%, people to 1) and returns the same dict as tip.

        text: Bill, optional tip percent, optional number of people.

    Example: math.tip_text("85.50 18% 2")
    Keywords: tip, gratuity, bill, split, restaurant, people, calculator
    Needs: none (pure function).
    """

def loan_text(text: str) -> Dict[str, str]:
    """Reads "principal rate% years [payments_per_year]" from text ("200000 6% 30") and returns the same dict as loan_payment: payment, total, interest and payments.

        text: Principal, annual rate percent, years, optional payments per year.

    Example: math.loan_text("200000 6% 30")
    Keywords: loan, payment, mortgage, monthly, emi, installment, calculator, interest
    Needs: none (pure function).
    """

def solve_quadratic(a: float, b: float, c: float) -> List[str]:
    """Solves a*x^2 + b*x + c = 0 and returns the roots as strings, real ones ascending (a double root once) and complex ones like "1+2i", "1-2i". With a=0 it solves the linear equation. x^2-3x+2 gives ["1", "2"].

        a: Coefficient of x^2.
        b: Coefficient of x.
        c: Constant term.

    Example: math.solve_quadratic(1, -3, 2)
    Task: solve a quadratic equation
    Keywords: roots of polynomial, quadratic formula, quadratic, equation, solve, roots, discriminant, parabola, complex, x2, formula
    Needs: none (pure function).
    """

def quadratic_text(text: str) -> Dict[str, str]:
    """Solves a quadratic written as text, either "x^2 - 3x + 2 = 0" or three coefficients "1 -3 2". Returns a, b, c, discriminant and the roots x1, x2 (complex roots as strings).

        text: Equation or coefficients a b c.

    Example: math.quadratic_text("x^2 - 3x + 2 = 0")
    Keywords: quadratic, equation, solve, roots, discriminant, parabola, x2, formula
    Needs: none (pure function).
    """

def fraction_calc(expr: str) -> str:
    """Evaluates an expression exactly with fractions: numbers, decimals, + - * / and parentheses ("3/4 + 1/6" gives "11/12"). Division is exact, never floating point.

        expr: Expression such as 3/4 + 1/6 or (1/2)*(2/3).

    Example: math.fraction_calc("3/4 + 1/6")
    Keywords: fraction, calculator, exact, rational, add, subtract, multiply, divide, frac
    Needs: none (pure function).
    """

def time_sum(expr: str) -> str:
    """Adds and subtracts clock times: "1:30 + 2:45" gives "4:15"; also h:mm:ss and unit forms "1h 30m + 45m". Hours may exceed 24; a negative total keeps its minus sign; seconds appear only when non-zero.

        expr: Terms joined by + and -, e.g. 1:30 + 2:45 - 0:20.

    Example: math.time_sum("1:30 + 2:45")
    Keywords: time, sum, add, subtract, hours, minutes, duration, clock, hh:mm, timesheet
    Needs: none (pure function).
    """

def ratio_simplify(ratio: str) -> str:
    """Reduces a ratio to the smallest whole numbers: "1920:1080" gives "16:9", "4:6:8" gives "2:3:4", "0.5:1.5" gives "1:3". Separators : / x , or spaces.

        ratio: Ratio such as 16:9.

    Example: math.ratio_simplify("1920:1080")
    Keywords: ratio, simplify, aspect, reduce, proportion, 16:9, screen
    Needs: none (pure function).
    """

def temperature_text(text: str) -> Dict[str, str]:
    """Converts text like "72f", "-40 C" or "300K" to all four scales and returns a dict with c, f, k and r.

        text: Number with a unit letter, e.g. 72f.

    Example: math.temperature_text("72f")
    Keywords: temperature, convert, celsius, fahrenheit, kelvin, rankine, degrees, weather
    Needs: none (pure function).
    """

def what_percent(a: float, b: float) -> float:
    """Returns what percent a is of b (what_percent(20, 50) = 40). b must not be zero.

        a: a value.
        b: b value.

    Example: math.what_percent(20, 50)
    Keywords: x is what percent of y, ratio as percent, what, percent, ratio, share, fraction, proportion
    Needs: none (pure function).
    """

def percent_change(old: float, new: float) -> float:
    """Returns the percentage change from old to new ((new-old)/|old|*100; 80 to 100 is 25). old must not be zero.

        old: old value.
        new: new value.

    Example: math.percent_change(80, 100)
    Task: percent change between two values
    Keywords: increase decrease percentage, growth rate, percent, change, increase, decrease, growth, difference, delta
    Needs: none (pure function).
    """

def add_percent(x: float, p: float) -> float:
    """Increases x by p percent (add_percent(200, 10) = 220).

        x: x value.
        p: p value.

    Example: math.add_percent(200, 10)
    Keywords: add, percent, increase, markup, plus, tax, raise
    Needs: none (pure function).
    """

def sub_percent(x: float, p: float) -> float:
    """Decreases x by p percent (sub_percent(200, 10) = 180).

        x: x value.
        p: p value.

    Example: math.sub_percent(200, 10)
    Keywords: subtract, percent, decrease, discount, minus, reduce
    Needs: none (pure function).
    """

def percent_diff(a: float, b: float) -> float:
    """Returns the symmetric percentage difference between a and b: |a-b| divided by their average (50 and 100 give 66.667). Both zero gives 0.

        a: a value.
        b: b value.

    Example: math.percent_diff(50, 100)
    Keywords: percent, difference, symmetric, relative, compare
    Needs: none (pure function).
    """

def markup(cost: float, price: float) -> float:
    """Returns the markup percentage of selling at price when the cost is cost: (price-cost)/cost*100. cost must not be zero.

        cost: cost value.
        price: price value.

    Example: math.markup(80, 100)
    Keywords: markup, percent, profit, cost, price, retail
    Needs: none (pure function).
    """

def margin(cost: float, price: float) -> float:
    """Returns the profit margin percentage of selling at price with cost cost: (price-cost)/price*100. price must not be zero.

        cost: cost value.
        price: price value.

    Example: math.margin(80, 100)
    Keywords: margin, percent, profit, gross, cost, price
    Needs: none (pure function).
    """

def price_from_markup(cost: float, markup_pct: float) -> float:
    """Returns the selling price for a cost and a markup percentage (cost 80, markup 25 gives 100).

        cost: cost value.
        markup_pct: markup_pct value.

    Example: math.price_from_markup(80, 25)
    Keywords: price, markup, cost, selling, retail
    Needs: none (pure function).
    """

def price_from_margin(cost: float, margin_pct: float) -> float:
    """Returns the selling price needed for a target profit margin percentage (cost 80, margin 20 gives 100). margin_pct must be below 100.

        cost: cost value.
        margin_pct: margin_pct value.

    Example: math.price_from_margin(80, 20)
    Keywords: price, margin, cost, selling, target, profit
    Needs: none (pure function).
    """

def tax_add(amount: float, rate: float) -> float:
    """Adds tax to a net amount: amount * (1 + rate/100) (tax_add(100, 20) = 120).

        amount: amount value.
        rate: rate value.

    Example: math.tax_add(100, 20)
    Keywords: tax, add, vat, gst, sales, gross, net, price
    Needs: none (pure function).
    """

def tax_remove(amount: float, rate: float) -> float:
    """Removes tax from a gross amount: amount / (1 + rate/100) (tax_remove(120, 20) = 100). rate must be above -100.

        amount: amount value.
        rate: rate value.

    Example: math.tax_remove(120, 20)
    Keywords: tax, remove, vat, gst, sales, net, gross, exclusive
    Needs: none (pure function).
    """

def reverse_percent(final: float, p: float, increase: bool = True) -> float:
    """Finds the original value before a percentage was added (increase=true) or taken off (increase=false): reverse_percent(120, 20) = 100, reverse_percent(80, 20, False) = 100.

        final: The value after the change.
        p: The percentage that was applied.
        increase: True if p was added, false if it was subtracted.

    Example: math.reverse_percent(120, 20)
    Keywords: reverse, percent, original, before, discount, undo, increase, find, price
    Needs: none (pure function).
    """

def markup_margin(cost: float, price: float) -> Dict[str, str]:
    """Returns profit, markup percent and margin percent for a cost and a selling price as a dict.

        cost: Cost.
        price: Selling price.

    Example: math.markup_margin(80, 100)
    Keywords: markup, margin, profit, cost, price, dict
    Needs: none (pure function).
    """

def discount(price: float, pct: float, pct2: float = 0.0) -> Dict[str, str]:
    """Applies a percentage discount (optionally a second stacked one) to a price and returns final price, amount saved and the effective percent as a dict.

        price: Original price.
        pct: Discount percent.
        pct2: Optional second discount percent applied after the first.

    Example: math.discount(80, 25)
    Task: price after discount
    Keywords: sale price, percent off, discount, sale, off, price, percent, saved, final, coupon
    Needs: none (pure function).
    """

def vat(amount: float, rate: float, inclusive: bool = False) -> Dict[str, str]:
    """Splits an amount into net, VAT and gross. With inclusive=false the amount is net and VAT is added; with inclusive=true the amount already includes VAT and is split out.

        amount: The amount.
        rate: VAT rate in percent.
        inclusive: The amount already includes VAT.

    Example: math.vat(100, 20)
    Keywords: sales tax, add tax, gst, vat, tax, split, net, gross, inclusive, exclusive, sales
    Needs: none (pure function).
    """

def tip(bill: float, pct: float = 15.0, people: int = 1) -> Dict[str, str]:
    """Calculates a tip and splits the bill: returns tip, total, per_person, tip_per_person and bill_per_person (rounded to cents) for a bill, a tip percent and a number of people.

        bill: Bill before tip.
        pct: Tip percent.
        people: Number of people sharing.

    Example: math.tip(85.5, 18, 2)
    Task: calculate a tip
    Keywords: gratuity, split the bill, tip, bill, split, restaurant, people, per, person
    Needs: none (pure function).
    """

def sum(values: List[float]) -> float:
    """Adds up a list of numbers with compensated summation (sum([0.1, 0.2, 0.3]) = 0.6). Empty list is an error.

        values: The numbers (a non-empty list of floats).

    Example: math.sum([1, 2, 3.5])
    Task: sum numbers
    Keywords: add up, total, sum, add, numbers, list
    Needs: none (pure function).
    """

def mean(values: List[float]) -> float:
    """Arithmetic mean (average) of a list of numbers.

        values: The numbers (a non-empty list of floats).

    Example: math.mean([2, 4, 9])
    Task: average of numbers
    Keywords: average, arithmetic mean, avg, mean, arithmetic, statistics
    Needs: none (pure function).
    """

def median(values: List[float]) -> float:
    """Median of a list: the middle value, or the mean of the two middle values for an even count.

        values: The numbers (a non-empty list of floats).

    Example: math.median([5, 1, 3, 2])
    Task: median of numbers
    Keywords: middle value, median, middle, value, statistics, 50th, percentile
    Needs: none (pure function).
    """

def mode(values: List[float]) -> List[float]:
    """Returns the most frequent value(s) in ascending order; several values are returned when they tie (all values if all are unique).

        values: The numbers (a non-empty list of floats).

    Example: math.mode([1, 2, 2, 3, 3])
    Keywords: mode, most, frequent, common, value, statistics, multimode
    Needs: none (pure function).
    """

def min(values: List[float]) -> float:
    """Smallest value in a list.

        values: The numbers (a non-empty list of floats).

    Example: math.min([4, 2, 9])
    Keywords: smallest, minimum, min, lowest
    Needs: none (pure function).
    """

def max(values: List[float]) -> float:
    """Largest value in a list.

        values: The numbers (a non-empty list of floats).

    Example: math.max([4, 2, 9])
    Keywords: largest, maximum, max, highest
    Needs: none (pure function).
    """

def range(values: List[float]) -> float:
    """Spread of a list: max minus min.

        values: The numbers (a non-empty list of floats).

    Example: math.range([4, 2, 9])
    Keywords: range, spread, max, min, difference
    Needs: none (pure function).
    """

def count(values: List[float]) -> int:
    """Number of values in a list.

        values: The numbers (a non-empty list of floats).

    Example: math.count([4, 2, 9])
    Keywords: count, length, number, of, values, size
    Needs: none (pure function).
    """

def geometric_mean(values: List[float]) -> float:
    """Geometric mean of positive numbers (the n-th root of their product, computed via logs). Zero or negative values are an error.

        values: The numbers (a non-empty list of floats).

    Example: math.geometric_mean([1, 4, 16])
    Keywords: geometric, mean, growth, average, statistics
    Needs: none (pure function).
    """

def harmonic_mean(values: List[float]) -> float:
    """Harmonic mean of positive numbers (count divided by the sum of reciprocals), suited to averaging rates. Zero or negative values are an error.

        values: The numbers (a non-empty list of floats).

    Example: math.harmonic_mean([40, 60])
    Keywords: harmonic, mean, rate, average, statistics
    Needs: none (pure function).
    """

def iqr(values: List[float]) -> float:
    """Interquartile range: the 75th minus the 25th percentile (linear interpolation).

        values: The numbers (a non-empty list of floats).

    Example: math.iqr([1, 2, 3, 4, 5, 6, 7, 8])
    Keywords: iqr, interquartile, range, quartile, spread, statistics
    Needs: none (pure function).
    """

def normalize(values: List[float]) -> List[float]:
    """Rescales values linearly to the range 0..1 (min-max normalisation). If all values are equal every result is 0.

        values: The numbers (a non-empty list of floats).

    Example: math.normalize([10, 20, 30])
    Keywords: normalize, min, max, scale, 0, 1, rescale, feature, scaling
    Needs: none (pure function).
    """

def cumulative_sum(values: List[float]) -> List[float]:
    """Running total: element i is the sum of the first i+1 values.

        values: The numbers (a non-empty list of floats).

    Example: math.cumulative_sum([1, 2, 3, 4])
    Keywords: cumulative, sum, running, total, cumsum, prefix
    Needs: none (pure function).
    """

def variance(values: List[float], sample: bool = True) -> float:
    """Variance of a list: sample variance (n-1 divisor, default) or population variance with sample=false. Sample variance needs at least 2 values.

        values: The numbers (a non-empty list of floats).
        sample: Use the sample (n-1) formula; false gives population variance.

    Example: math.variance([2, 4, 4, 4, 5, 5, 7, 9], sample=False)
    Keywords: variance, spread, statistics, sample, population
    Needs: none (pure function).
    """

def stdev(values: List[float], sample: bool = True) -> float:
    """Standard deviation: sample (n-1 divisor, default) or population with sample=false. For [2,4,4,4,5,5,7,9] the population value is 2.

        values: The numbers (a non-empty list of floats).
        sample: Use the sample (n-1) formula; false gives population stdev.

    Example: math.stdev([2, 4, 4, 4, 5, 5, 7, 9], sample=False)
    Task: standard deviation
    Keywords: spread, sigma, standard, deviation, stdev, statistics, sample, population
    Needs: none (pure function).
    """

def zscores(values: List[float], sample: bool = False) -> List[float]:
    """Standardises values: (x - mean) / stdev for each value. Uses the population stdev unless sample=true. All-equal data is an error (stdev is 0).

        values: The numbers (a non-empty list of floats).
        sample: Use the sample stdev.

    Example: math.zscores([1, 2, 3, 4, 5])
    Keywords: zscore, z, score, standardize, standard, statistics
    Needs: none (pure function).
    """

def quantile(values: List[float], q: float) -> float:
    """Quantile q (0..1) with linear interpolation between ranks (NumPy default): quantile(v, 0.5) is the median.

        values: The numbers (a non-empty list of floats).
        q: Fraction between 0 and 1.

    Example: math.quantile([1, 2, 3, 4], 0.25)
    Keywords: quantile, quartile, percentile, interpolation, statistics
    Needs: none (pure function).
    """

def percentile(values: List[float], p: float) -> float:
    """Percentile p (0..100) with linear interpolation: percentile(v, 90) is the value below which 90 percent of the data fall.

        values: The numbers (a non-empty list of floats).
        p: Percent between 0 and 100.

    Example: math.percentile([1, 2, 3, 4, 5], 90)
    Keywords: percentile, rank, statistics, interpolation
    Needs: none (pure function).
    """

def moving_average(values: List[float], window: int = 3) -> List[float]:
    """Simple moving average with the given window; returns len(values)-window+1 averages (only full windows). window must be 1..len(values).

        values: The numbers (a non-empty list of floats).
        window: Window size.

    Example: math.moving_average([1, 2, 3, 4, 5], 3)
    Keywords: moving, average, rolling, mean, smoothing, sma, window
    Needs: none (pure function).
    """

def outliers(values: List[float], k: float = 1.5) -> List[float]:
    """Returns the values outside the Tukey fences Q1 - k*IQR and Q3 + k*IQR (k defaults to 1.5), in input order.

        values: The numbers (a non-empty list of floats).
        k: Fence multiplier.

    Example: math.outliers([1, 2, 2, 3, 3, 4, 100])
    Keywords: outliers, anomalies, tukey, iqr, fences, extreme, values
    Needs: none (pure function).
    """

def correlation(xs: List[float], ys: List[float]) -> float:
    """Pearson correlation coefficient of two equal-length lists (-1..1). Needs at least 2 points and non-constant data.

        xs: X values.
        ys: Y values.

    Example: math.correlation([1, 2, 3], [2, 4, 6])
    Keywords: correlation, pearson, r, statistics, relationship
    Needs: none (pure function).
    """

def linear_regression(xs: List[float], ys: List[float]) -> Dict[str, str]:
    """Least-squares line y = slope*x + intercept through the points; returns slope, intercept, r and r2 as a dict. xs must not all be equal.

        xs: X values.
        ys: Y values.

    Example: math.linear_regression([1, 2, 3], [2, 4, 6])
    Keywords: linear, regression, least, squares, trend, line, fit, slope, intercept, r2
    Needs: none (pure function).
    """

def round_to(x: float, decimals: int = 0, mode: str = "half_up") -> float:
    """Rounds x to a number of decimals (negative rounds to tens, hundreds) using its decimal form, so 2.675 gives 2.68 with half_up. Modes: half_up (default, away from zero on ties), half_even (banker's), half_down, floor, ceil, trunc.

        x: Number to round.
        decimals: Decimal places (-300..300).
        mode: Rounding mode. One of: half_up, half_even, half_down, floor, ceil, trunc.

    Example: math.round_to(2.675, 2)
    Task: round a number
    Keywords: decimal places, rounding, round, decimals, half, even, bankers, floor, ceil, truncate, places
    Needs: none (pure function).
    """

def round_to_multiple(x: float, multiple: float, mode: str = "half_up") -> float:
    """Rounds x to the nearest multiple of m (round_to_multiple(17, 5) = 15, (18, 5) = 20, 0.1 steps work exactly). Same modes as round_to. m must not be zero.

        x: Number to round.
        multiple: The step.
        mode: Rounding mode. One of: half_up, half_even, half_down, floor, ceil, trunc.

    Example: math.round_to_multiple(18, 5)
    Keywords: round, nearest, multiple, step, increment, quantize, snap
    Needs: none (pure function).
    """

def clamp(x: float, lo: float, hi: float) -> float:
    """Limits x to the range lo..hi (clamp(15, 0, 10) = 10). lo must not exceed hi.

        x: Value.
        lo: Lower bound.
        hi: Upper bound.

    Example: math.clamp(15, 0, 10)
    Keywords: limit to range, bound a value, min max constrain, clamp, limit, bound, min, max, constrain, range
    Needs: none (pure function).
    """

def lerp(a: float, b: float, t: float) -> float:
    """Linear interpolation a + (b-a)*t (lerp(10, 20, 0.25) = 12.5); t outside 0..1 extrapolates.

        a: Start.
        b: End.
        t: Position, 0 = a, 1 = b.

    Example: math.lerp(10, 20, 0.25)
    Keywords: lerp, linear, interpolation, mix, blend, between
    Needs: none (pure function).
    """

def inverse_lerp(a: float, b: float, x: float) -> float:
    """Returns where x sits between a and b as a fraction ((x-a)/(b-a)). a and b must differ.

        a: Start.
        b: End.
        x: Value.

    Example: math.inverse_lerp(10, 20, 12.5)
    Keywords: inverse, lerp, unlerp, fraction, position, progress
    Needs: none (pure function).
    """

def map_range(x: float, in_lo: float, in_hi: float, out_lo: float, out_hi: float) -> float:
    """Maps x from the range in_lo..in_hi to out_lo..out_hi linearly (map_range(5, 0, 10, 0, 100) = 50). in_lo and in_hi must differ; values outside are extrapolated.

        x: Value.
        in_lo: Input range start.
        in_hi: Input range end.
        out_lo: Output range start.
        out_hi: Output range end.

    Example: math.map_range(5, 0, 10, 0, 100)
    Keywords: rescale, remap value between ranges, map, range, remap, scale, convert, interval
    Needs: none (pure function).
    """

def solve_linear(a: float, b: float) -> float:
    """Solves a*x + b = 0 for x (solve_linear(2, -6) = 3). a must not be zero.

        a: Coefficient of x.
        b: Constant term.

    Example: math.solve_linear(2, -6)
    Keywords: solve, linear, equation, ax+b, root, x
    Needs: none (pure function).
    """

def solve_linear_system_2x2(matrix: List[float]) -> List[float]:
    """Solves two linear equations in x and y given as the augmented matrix [a1, b1, c1, a2, b2, c2] meaning a1*x+b1*y=c1 and a2*x+b2*y=c2. Returns [x, y]; singular systems are an error.

        matrix: Six numbers: a1, b1, c1, a2, b2, c2.

    Example: math.solve_linear_system_2x2([1, 1, 3, 1, -1, 1])
    Keywords: linear, system, equations, simultaneous, 2x2, solve, cramer, gauss
    Needs: none (pure function).
    """

def solve_linear_system_3x3(matrix: List[float]) -> List[float]:
    """Solves three linear equations in x, y, z given as the row-by-row augmented matrix (12 numbers: a, b, c, d for each equation a*x+b*y+c*z=d). Returns [x, y, z]; singular systems are an error.

        matrix: Twelve numbers, four per equation.

    Example: math.solve_linear_system_3x3([1,1,1,6, 0,2,5,-4, 2,5,-1,27])
    Keywords: linear, system, equations, simultaneous, 3x3, solve, gauss, elimination
    Needs: none (pure function).
    """

def simplify_fraction(n: int, d: int) -> str:
    """Reduces n/d to lowest terms with a positive denominator (simplify_fraction(6, -8) = "-3/4"; whole results print without a denominator). d must not be zero.

        n: Numerator.
        d: Denominator.

    Example: math.simplify_fraction(6, 8)
    Task: simplify a fraction
    Keywords: reduce fraction, lowest terms, simplify, fraction, reduce, lowest, terms, gcd, numerator, denominator
    Needs: none (pure function).
    """

def fraction_add(a: str, b: str) -> str:
    """Exact fraction addition of a and b (each "3/4", "0.75", "2" or "1 1/2"). The result is in lowest terms, printed as n/d or a whole number.

        a: First fraction.
        b: Second fraction.

    Example: math.fraction_add("1/2", "1/3")
    Keywords: fraction, addition, exact, rational, arithmetic
    Needs: none (pure function).
    """

def fraction_div(a: str, b: str) -> str:
    """Exact fraction division of a and b (each "3/4", "0.75", "2" or "1 1/2"). The result is in lowest terms, printed as n/d or a whole number.

        a: First fraction.
        b: Second fraction.

    Example: math.fraction_div("1/2", "1/3")
    Keywords: fraction, division, exact, rational, arithmetic
    Needs: none (pure function).
    """

def fraction_mul(a: str, b: str) -> str:
    """Exact fraction multiplication of a and b (each "3/4", "0.75", "2" or "1 1/2"). The result is in lowest terms, printed as n/d or a whole number.

        a: First fraction.
        b: Second fraction.

    Example: math.fraction_mul("1/2", "1/3")
    Keywords: fraction, multiplication, exact, rational, arithmetic
    Needs: none (pure function).
    """

def fraction_sub(a: str, b: str) -> str:
    """Exact fraction subtraction of a and b (each "3/4", "0.75", "2" or "1 1/2"). The result is in lowest terms, printed as n/d or a whole number.

        a: First fraction.
        b: Second fraction.

    Example: math.fraction_sub("1/2", "1/3")
    Keywords: fraction, subtraction, exact, rational, arithmetic
    Needs: none (pure function).
    """

def decimal_to_fraction(x: str, max_denominator: int = 0) -> str:
    """Converts a decimal to a fraction. With max_denominator=0 (default) the conversion is exact for the decimal text (0.125 = 1/8, 0.333 = 333/1000); a positive max_denominator finds the closest fraction up to that denominator (3.14159 with 1000 = 355/113).

        x: Decimal number or fraction.
        max_denominator: Largest denominator (0 = exact).

    Example: math.decimal_to_fraction("0.125")
    Keywords: 0.75 as fraction, float to fraction, decimal, to, fraction, convert, rational, numerator, denominator
    Needs: none (pure function).
    """

def mixed_number(x: str) -> str:
    """Writes an improper fraction or decimal as a mixed number ("7/2" gives "3 1/2", 2.75 gives "2 3/4", -7/2 gives "-3 1/2").

        x: Fraction like 7/2 or a decimal.

    Example: math.mixed_number("7/2")
    Keywords: mixed, number, improper, fraction, whole, part, convert
    Needs: none (pure function).
    """

def scale_ratio(ratio: str, known: float, position: int = 1) -> str:
    """Scales a ratio so that one part has a known value and returns all parts: scale_ratio("16:9", 1920) gives "1920:1080"; position picks which part (1-based) the value belongs to.

        ratio: Ratio such as 16:9.
        known: Value of the chosen part.
        position: 1-based index of the part that equals known.

    Example: math.scale_ratio("16:9", 1920)
    Keywords: scale, ratio, aspect, resize, proportion, width, height, dimension
    Needs: none (pure function).
    """

def temperature(value: float, from_unit: str, to_unit: str) -> float:
    """Converts a temperature between c (Celsius), f (Fahrenheit), k (Kelvin) and r (Rankine); full unit names and a degree sign are accepted. Values below absolute zero are an error.

        value: Temperature.
        from_unit: Source unit: c, f, k or r.
        to_unit: Target unit: c, f, k or r.

    Example: math.temperature(100, "c", "f")
    Task: convert temperature
    Keywords: celsius fahrenheit kelvin, degrees conversion, temperature, convert, celsius, fahrenheit, kelvin, rankine, degrees
    Needs: none (pure function).
    """

def degrees_to_radians(degrees: float) -> float:
    """Converts degrees to radians (180 gives 3.14159...).

        degrees: Angle in degrees.

    Example: math.degrees_to_radians(180)
    Keywords: degrees, radians, convert, angle, deg, rad
    Needs: none (pure function).
    """

def radians_to_degrees(radians: float) -> float:
    """Converts radians to degrees (pi gives 180).

        radians: Angle in radians.

    Example: math.radians_to_degrees(3.141592653589793)
    Keywords: radians, degrees, convert, angle, deg, rad
    Needs: none (pure function).
    """

def polar_to_cartesian(r: float, theta: float, degrees: bool = False) -> Dict[str, str]:
    """Converts polar (r, theta) to Cartesian x and y. theta is in radians unless degrees=true. Returns a dict with x and y (tiny floating-point noise below 1e-12 is cleaned to 0).

        r: Radius.
        theta: Angle.
        degrees: theta is in degrees.

    Example: math.polar_to_cartesian(2, 90, degrees=True)
    Keywords: polar, cartesian, convert, coordinates, r, theta, x, y
    Needs: none (pure function).
    """

def cartesian_to_polar(x: float, y: float, degrees: bool = False) -> Dict[str, str]:
    """Converts Cartesian (x, y) to polar: returns r and theta (radians, or degrees with degrees=true, in the range -pi..pi / -180..180).

        x: X.
        y: Y.
        degrees: Return theta in degrees.

    Example: math.cartesian_to_polar(1, 1, degrees=True)
    Keywords: cartesian, polar, convert, coordinates, r, theta, x, y
    Needs: none (pure function).
    """

def distance_2d(x1: float, y1: float, x2: float, y2: float) -> float:
    """Euclidean distance between (x1, y1) and (x2, y2).

        x1: X1.
        y1: Y1.
        x2: X2.
        y2: Y2.

    Example: math.distance_2d(0, 0, 3, 4)
    Keywords: distance, 2d, euclidean, points, length, between
    Needs: none (pure function).
    """

def distance_3d(x1: float, y1: float, z1: float, x2: float, y2: float, z2: float) -> float:
    """Euclidean distance between (x1, y1, z1) and (x2, y2, z2).

        x1: X1.
        y1: Y1.
        z1: Z1.
        x2: X2.
        y2: Y2.
        z2: Z2.

    Example: math.distance_3d(0, 0, 0, 1, 2, 2)
    Keywords: distance, 3d, euclidean, points, length, between
    Needs: none (pure function).
    """

def haversine(lat1: float, lon1: float, lat2: float, lon2: float) -> Dict[str, str]:
    """Great-circle distance between two latitude/longitude points (degrees) using the haversine formula on a 6371.0088 km sphere. Returns km, mi, nmi and the initial bearing in degrees. Latitudes must be -90..90 and longitudes -180..180.

        lat1: Latitude 1.
        lon1: Longitude 1.
        lat2: Latitude 2.
        lon2: Longitude 2.

    Example: math.haversine(51.5074, -0.1278, 48.8566, 2.3522)
    Task: distance between coordinates
    Keywords: great circle distance, latitude longitude distance, haversine, distance, great, circle, latitude, longitude, gps, km, miles, bearing, geo
    Needs: none (pure function).
    """

def geometry(shape: str, a: float = 0.0, b: float = 0.0, c: float = 0.0) -> Dict[str, str]:
    """Area, perimeter, volume and other measures of a shape as a dict. circle(a=radius), rectangle(a=width, b=height), square(a), triangle(a, b, c = three sides), ellipse(a, b = semi-axes), sphere(a=radius), cube(a), cylinder(a=radius, b=height), cone(a=radius, b=height). Unused dimensions stay 0; non-positive needed ones are an error.

        shape: Shape name.
        a: First dimension.
        b: Second dimension.
        c: Third dimension.

    Example: math.geometry("cylinder", 2, 5)
    Keywords: geometry, shape, area, volume, perimeter, circumference, surface, circle, rectangle, triangle, sphere, cylinder, cone, cube
    Needs: none (pure function).
    """

def circle_area(r: float) -> float:
    """Area of a circle: pi*r^2.

        r: Radius.

    Example: math.circle_area(3)
    Keywords: area of circle, pi r squared, geometry, area, volume, perimeter, surface, shape, size, circle
    Needs: none (pure function).
    """

def circle_circumference(r: float) -> float:
    """Circumference of a circle: 2*pi*r.

        r: Radius.

    Example: math.circle_circumference(3)
    Keywords: geometry, area, volume, perimeter, surface, shape, size, circle, circumference
    Needs: none (pure function).
    """

def rect_area(w: float, h: float) -> float:
    """Area of a rectangle: w*h.

        w: Width.
        h: Height.

    Example: math.rect_area(3, 4)
    Keywords: geometry, area, volume, perimeter, surface, shape, size, rect
    Needs: none (pure function).
    """

def rect_perimeter(w: float, h: float) -> float:
    """Perimeter of a rectangle: 2*(w+h).

        w: Width.
        h: Height.

    Example: math.rect_perimeter(3, 4)
    Keywords: geometry, area, volume, perimeter, surface, shape, size, rect
    Needs: none (pure function).
    """

def triangle_area(a: float, b: float, c: float) -> float:
    """Area of a triangle from its three side lengths (Heron's formula; 3, 4, 5 gives 6). Sides that violate the triangle inequality are an error.

        a: Side a.
        b: Side b.
        c: Side c.

    Example: math.triangle_area(3, 4, 5)
    Keywords: geometry, area, volume, perimeter, surface, shape, size, triangle
    Needs: none (pure function).
    """

def triangle_perimeter(a: float, b: float, c: float) -> float:
    """Perimeter of a triangle: a+b+c (sides must satisfy the triangle inequality).

        a: Side a.
        b: Side b.
        c: Side c.

    Example: math.triangle_perimeter(3, 4, 5)
    Keywords: geometry, area, volume, perimeter, surface, shape, size, triangle
    Needs: none (pure function).
    """

def sphere_volume(r: float) -> float:
    """Volume of a sphere: 4/3*pi*r^3.

        r: Radius.

    Example: math.sphere_volume(2)
    Keywords: geometry, area, volume, perimeter, surface, shape, size, sphere
    Needs: none (pure function).
    """

def sphere_area(r: float) -> float:
    """Surface area of a sphere: 4*pi*r^2.

        r: Radius.

    Example: math.sphere_area(2)
    Keywords: geometry, area, volume, perimeter, surface, shape, size, sphere
    Needs: none (pure function).
    """

def cylinder_volume(r: float, h: float) -> float:
    """Volume of a cylinder: pi*r^2*h.

        r: Radius.
        h: Height.

    Example: math.cylinder_volume(2, 5)
    Keywords: geometry, area, volume, perimeter, surface, shape, size, cylinder
    Needs: none (pure function).
    """

def cylinder_area(r: float, h: float) -> float:
    """Total surface area of a cylinder: 2*pi*r*(r+h).

        r: Radius.
        h: Height.

    Example: math.cylinder_area(2, 5)
    Keywords: geometry, area, volume, perimeter, surface, shape, size, cylinder
    Needs: none (pure function).
    """

def cone_volume(r: float, h: float) -> float:
    """Volume of a cone: pi*r^2*h/3.

        r: Radius.
        h: Height.

    Example: math.cone_volume(2, 5)
    Keywords: geometry, area, volume, perimeter, surface, shape, size, cone
    Needs: none (pure function).
    """

def cone_area(r: float, h: float) -> float:
    """Total surface area of a cone: pi*r*(r+slant) with slant = sqrt(r^2+h^2).

        r: Radius.
        h: Height.

    Example: math.cone_area(2, 5)
    Keywords: geometry, area, volume, perimeter, surface, shape, size, cone
    Needs: none (pure function).
    """

def pythagoras(a: float = 0.0, b: float = 0.0, c: float = 0.0) -> float:
    """Solves a right triangle with legs a, b and hypotenuse c: set exactly one of them to 0 (the unknown) and the other two to their values. pythagoras(3, 4, 0) = 5; pythagoras(0, 12, 13) = 5. The hypotenuse must be longer than a leg.

        a: Leg a (0 if unknown).
        b: Leg b (0 if unknown).
        c: Hypotenuse c (0 if unknown).

    Example: math.pythagoras(3, 4, 0)
    Keywords: pythagoras, pythagorean, theorem, right, triangle, hypotenuse, leg, a2+b2=c2
    Needs: none (pure function).
    """

def compound_interest(principal: float, rate: float, years: float, n: int = 1) -> Dict[str, str]:
    """Future value with compound interest: principal * (1 + rate/n)^(n*years), rate in percent per year, n compounding periods per year (n=1 annual, 12 monthly, 0 continuous). Returns amount, interest and the effective annual rate as a dict.

        principal: Starting amount.
        rate: Annual rate in percent.
        years: Duration in years.
        n: Compounding periods per year (0 = continuous).

    Example: math.compound_interest(1000, 5, 10, 1)
    Keywords: interest over time, savings growth, compound, interest, investment, savings, growth, apy, future, value, compounding
    Needs: none (pure function).
    """

def simple_interest(principal: float, rate: float, years: float) -> Dict[str, str]:
    """Simple (non-compounding) interest: principal * rate/100 * years. Returns interest and total as a dict.

        principal: Amount.
        rate: Annual rate in percent.
        years: Duration in years.

    Example: math.simple_interest(1000, 5, 3)
    Keywords: simple, interest, loan, deposit, principal, rate, years
    Needs: none (pure function).
    """

def amortization_summary(principal: float, annual_rate: float, years: float, per_year: int = 12, after: int = 12) -> Dict[str, str]:
    """Splits a loan into interest and principal: after `after` payments returns interest_paid, principal_paid, balance, first_interest, first_principal, payment and payments_left (same inputs as loan_payment).

        principal: Amount borrowed.
        annual_rate: Annual rate in percent.
        years: Term in years.
        per_year: Payments per year.
        after: Number of payments already made.

    Example: math.amortization_summary(200000, 6, 30, 12, 12)
    Keywords: amortization, schedule, loan, balance, principal, interest, split, remaining, mortgage
    Needs: none (pure function).
    """

def future_value(present: float, rate: float, periods: float, payment: float = 0.0) -> float:
    """Future value of a starting amount plus optional regular payments at the end of each period: present*(1+r)^n + payment*((1+r)^n-1)/r, with rate the percent per period.

        present: Starting amount.
        rate: Rate per period in percent.
        periods: Number of periods.
        payment: Payment added at the end of every period.

    Example: math.future_value(1000, 5, 10, 100)
    Keywords: future, value, fv, investment, annuity, savings, growth, periodic, payments
    Needs: none (pure function).
    """

def present_value(future: float, rate: float, periods: float, payment: float = 0.0) -> float:
    """Present value of a future amount plus optional regular payments, discounted at a percent rate per period: future/(1+r)^n + payment*(1-(1+r)^-n)/r.

        future: Future amount.
        rate: Discount rate per period in percent.
        periods: Number of periods.
        payment: Payment received at the end of every period.

    Example: math.present_value(1628.89, 5, 10)
    Keywords: present, value, pv, discount, npv, annuity, worth, today
    Needs: none (pure function).
    """

def cagr(start: float, end: float, years: float) -> float:
    """Compound annual growth rate in percent from a start value to an end value over some years ((end/start)^(1/years)-1). Start and end must be positive; years > 0.

        start: Start value.
        end: End value.
        years: Years.

    Example: math.cagr(1000, 2000, 5)
    Keywords: cagr, compound, annual, growth, rate, return, investment, annualized
    Needs: none (pure function).
    """

def roi(initial: float, final: float) -> float:
    """Return on investment in percent: (final - initial) / initial * 100. initial must not be zero.

        initial: Amount invested.
        final: Value at the end.

    Example: math.roi(1000, 1250)
    Keywords: roi, return, on, investment, profit, percent, gain
    Needs: none (pure function).
    """

def rule_of_72(rate: float) -> Dict[str, str]:
    """Years to double money at an annual percent rate: returns the rule-of-72 estimate and the exact value ln(2)/ln(1+r) as a dict. rate must be positive.

        rate: Annual rate in percent.

    Example: math.rule_of_72(6)
    Keywords: rule, of, 72, doubling, time, investment, years, to, double
    Needs: none (pure function).
    """

def break_even(fixed_costs: float, price: float, variable_cost: float) -> Dict[str, str]:
    """Break-even point: units = fixed_costs / (price - variable_cost). Returns exact units, units rounded up and the revenue at break-even. price must exceed variable_cost.

        fixed_costs: Fixed costs.
        price: Selling price per unit.
        variable_cost: Variable cost per unit.

    Example: math.break_even(5000, 25, 15)
    Keywords: break, even, point, units, fixed, costs, contribution, margin, business
    Needs: none (pure function).
    """

def unit_price(price: float, qty: float) -> float:
    """Price per unit: price / quantity (quantity must be positive).

        price: Total price.
        qty: Quantity.

    Example: math.unit_price(3.5, 12)
    Keywords: unit, price, per, cost, item
    Needs: none (pure function).
    """

def compare_unit_price(a_price: float, a_qty: float, b_price: float, b_qty: float) -> Dict[str, str]:
    """Compares two offers by unit price: returns each unit price, which is cheaper (a, b or same) and the saving percent of the cheaper one.

        a_price: Price of offer A.
        a_qty: Quantity of offer A.
        b_price: Price of offer B.
        b_qty: Quantity of offer B.

    Example: math.compare_unit_price(3.5, 12, 5, 20)
    Keywords: compare, unit, price, better, deal, cheaper, shopping, value
    Needs: none (pure function).
    """

def average_speed(distance: float, time: float) -> float:
    """Average speed = distance / time (any consistent units). time must be positive.

        distance: distance.
        time: time.

    Example: math.average_speed(150, 2.5)
    Keywords: speed, distance, time, velocity, travel, average, pace
    Needs: none (pure function).
    """

def travel_time(distance: float, speed: float) -> float:
    """Time = distance / speed. speed must be positive.

        distance: distance.
        speed: speed.

    Example: math.travel_time(150, 60)
    Keywords: speed, distance, time, velocity, travel, average, pace
    Needs: none (pure function).
    """

def travel_distance(speed: float, time: float) -> float:
    """Distance = speed * time.

        speed: Speed.
        time: Time.

    Example: math.travel_distance(60, 2.5)
    Keywords: distance, speed, time, travel, how, far
    Needs: none (pure function).
    """

def bmi(kg: float, m: float) -> Dict[str, str]:
    """Body mass index = kg / m^2, with the WHO category (underweight, normal, overweight, obese). Weight in kilograms, height in metres; both must be positive.

        kg: Weight in kilograms.
        m: Height in metres.

    Example: math.bmi(70, 1.75)
    Keywords: body mass index, bmi, body, mass, index, weight, height, health
    Needs: none (pure function).
    """

def angle_sum(expr: str) -> str:
    """Adds and subtracts angles written as degrees/minutes/seconds: "30°15'20\" + 45°50'" gives 76°05'20". Each term accepts 40°26'46"N style, 40 26 46 or plain degrees. Result is in d°mm'ss" form.

        expr: Angles joined by + and -.

    Example: math.angle_sum("30 15 20 + 45 50")
    Keywords: angle, sum, add, subtract, degrees, minutes, seconds, dms
    Needs: none (pure function).
    """

def dms_to_decimal(dms: str) -> float:
    """Converts degrees/minutes/seconds (40°26'46"N, 40 26 46, -73°59'W) to decimal degrees; S and W (or a leading minus) make it negative.

        dms: Angle text.

    Example: math.dms_to_decimal("40°26'46\"N")
    Keywords: dms, degrees, minutes, seconds, decimal, latitude, longitude, gps, coordinates
    Needs: none (pure function).
    """

def decimal_to_dms(degrees: float) -> str:
    """Converts decimal degrees to d°mm'ss.ss" form (40.446 gives 40°26'45.6"); the sign is kept as a minus.

        degrees: Decimal degrees.

    Example: math.decimal_to_dms(40.446)
    Keywords: decimal, degrees, dms, minutes, seconds, latitude, longitude, gps, coordinates
    Needs: none (pure function).
    """

def combinations(n: int, k: int) -> str:
    """Number of ways to choose k items from n without order, C(n, k), as a decimal string (combinations(52, 5) = 2598960). k > n gives 0. n is limited to 0..20000.

        n: Total items.
        k: Items chosen.

    Example: math.combinations(52, 5)
    Keywords: n choose k, binomial coefficient, combinations, choose, binomial, coefficient, ncr, n, k, counting
    Needs: none (pure function).
    """

def permutations(n: int, k: int) -> str:
    """Number of ordered arrangements of k items from n, P(n, k) = n!/(n-k)!, as a decimal string (permutations(10, 3) = 720). k > n gives 0. n is limited to 0..20000.

        n: Total items.
        k: Items arranged.

    Example: math.permutations(10, 3)
    Keywords: permutations, arrangements, npr, ordered, counting
    Needs: none (pure function).
    """

def binomial_probability(n: int, k: int, p: float) -> float:
    """Probability of exactly k successes in n independent trials with success probability p: C(n,k) p^k (1-p)^(n-k). n up to 10 million; p in 0..1.

        n: Trials.
        k: Successes.
        p: Success probability, 0..1.

    Example: math.binomial_probability(10, 5, 0.5)
    Keywords: binomial, probability, distribution, trials, successes, coin, flips, statistics, pmf
    Needs: none (pure function).
    """

def normal_cdf(x: float, mean: float = 0.0, sd: float = 1.0) -> float:
    """Cumulative probability P(X <= x) of a normal distribution with the given mean (default 0) and standard deviation sd (default 1, must be positive). normal_cdf(1.96) is about 0.975.

        x: Value.
        mean: Mean.
        sd: Standard deviation.

    Example: math.normal_cdf(1.96)
    Keywords: normal, distribution, cdf, gaussian, probability, bell, curve, z, score, cumulative
    Needs: none (pure function).
    """

def normal_pdf(x: float, mean: float = 0.0, sd: float = 1.0) -> float:
    """Probability density of a normal distribution at x (mean default 0, sd default 1, sd must be positive).

        x: Value.
        mean: Mean.
        sd: Standard deviation.

    Example: math.normal_pdf(0)
    Keywords: normal, distribution, pdf, gaussian, density, bell, curve
    Needs: none (pure function).
    """

def dice_probability(sides: int, count: int, target: int) -> float:
    """Probability that `count` dice with `sides` sides sum to exactly `target` (two six-sided dice summing to 7 gives 1/6). count*sides is limited to 5000.

        sides: Sides per die (>= 2).
        count: Number of dice (>= 1).
        target: Target sum.

    Example: math.dice_probability(6, 2, 7)
    Keywords: dice, probability, sum, roll, odds, die, sides, gaming
    Needs: none (pure function).
    """
