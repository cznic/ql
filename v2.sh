grep -v FileV2 $1 > old
grep FileV2 $1 | sed s/FileV2/File/ > new
benchcmp -mag -changed old new | sed s/Benchmark//
rm -f old new
