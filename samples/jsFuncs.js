const add = (a, b) => {
    // Add two numbers together
    const sum = a + b;
    return sum;
};

const pow = (a, b) => {
    // Take one number ot the power of another
    const power = a**b
    return power;
};

const sub = (a, b) => {
    // Subtract a number from another number
    const diff = a - b;
    return diff;
};

const mul = (a, b) => {
    // Multiply two numbers together
    const product = a * b;
    return product;
};

const less = (a, b) => {
    // Find the lesser of two numbers
    const small = Math.min(a, b);
    return small;
};

const more = (a, b) => {
    // Find the greater of two numbers
    const big = Math.max(a, b);
    return big;
};

const div = (a, b) => {
    // Divide a number by another number
    const ratio = a / b;
    return ratio;
};
