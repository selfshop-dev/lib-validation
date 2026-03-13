BEGIN { in_section=0; sec_idx=0 }
/^# ===/ {
    sec_idx++; line = $0; gsub(/# === | ===/, "", line);
    section_title[sec_idx] = line; section_max[sec_idx] = 0;
    cur_sec = sec_idx; in_section=1
}
/^[a-zA-Z_-]+:.*##/ {
    if (in_section) {
        split($0, parts, ":.*?## ");
        name = parts[1]; desc = parts[2];
        target_sec[NR] = cur_sec; target_name[NR] = name; target_desc[NR] = desc;
        if (length(name) > section_max[cur_sec]) section_max[cur_sec] = length(name)
    }
}
END {
    for (i = 1; i <= NR; i++) {
        if (i in target_sec) {
            s = target_sec[i];
            if (!(s in printed_sec)) { printf "\n\033[1m%s:\033[0m\n", section_title[s]; printed_sec[s] = 1 }
            fmt = sprintf("  \033[36m%%-%ds\033[0m %%s\n", section_max[s] + 2);
            printf fmt, target_name[i], target_desc[i]
        }
    }
}